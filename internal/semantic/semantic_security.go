package semantic

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// isInHTTPHandler returns true when the current function is an HTTP handler.
// Detected by the presence of an http.ResponseWriter parameter.
func (a *Analyzer) isInHTTPHandler() bool {
	if a.currentFunc == nil {
		return false
	}
	for _, param := range a.currentFunc.Parameters {
		if named, ok := param.Type.(*ast.NamedType); ok {
			if named.Name == "http.ResponseWriter" {
				return true
			}
		}
	}
	return false
}

// checkSQLInterpolation detects string interpolation in SQL query arguments
// to pg.Query, pg.QueryRow, pg.Exec and their Tx variants. This catches a
// class of SQL injection where Kukicha's "{var}" syntax interpolates user
// data into the query string before pgx's parameterization can protect it.
func (a *Analyzer) checkSQLInterpolation(qualifiedName string, expr *ast.MethodCallExpr, pipedArg *TypeInfo) {
	// Functions where the SQL string is an argument
	sqlFunctions := map[string]bool{
		"pg.Query":      true,
		"pg.QueryRow":   true,
		"pg.Exec":       true,
		"pg.TxQuery":    true,
		"pg.TxQueryRow": true,
		"pg.TxExec":     true,
	}

	if !sqlFunctions[qualifiedName] {
		return
	}

	// Determine the index of the SQL string argument.
	// Normal call: pg.Query(pool, "SELECT ...", args) → SQL at index 1
	// Piped call:  pool |> pg.Query("SELECT ...", args) → SQL at index 0
	//   (pipe inserts pool as first arg at codegen; AST Arguments only has explicit args)
	sqlArgIndex := 1
	if pipedArg != nil {
		sqlArgIndex = 0
	}

	if sqlArgIndex >= len(expr.Arguments) {
		return
	}

	sqlArg := expr.Arguments[sqlArgIndex]
	if strLit, ok := sqlArg.(*ast.StringLiteral); ok && strLit.Interpolated {
		a.error(strLit.Pos(), fmt.Sprintf(
			"SQL injection risk: string interpolation in %s query — use parameter placeholders ($1, $2, ...) instead",
			qualifiedName,
		))
	}
}

// checkHTMLNonLiteral warns when http.HTML (or its alias) is called with a
// non-literal content argument, which is a direct XSS vector.
func (a *Analyzer) checkHTMLNonLiteral(qualifiedName string, expr *ast.MethodCallExpr, pipedArg *TypeInfo) {
	htmlFunctions := map[string]bool{
		"httphelper.HTML": true,
		"http.HTML":       true,
	}
	if !htmlFunctions[qualifiedName] {
		return
	}

	// Content is the second arg (index 1) in a plain call, or the first (index 0)
	// when the ResponseWriter is piped in.
	contentArgIndex := 1
	if pipedArg != nil {
		contentArgIndex = 0
	}
	if contentArgIndex >= len(expr.Arguments) {
		return
	}

	contentArg := expr.Arguments[contentArgIndex]
	if _, ok := contentArg.(*ast.StringLiteral); !ok {
		a.error(expr.Pos(), fmt.Sprintf(
			"XSS risk: %s with non-literal content — use http.SafeHTML to HTML-escape user-controlled content",
			qualifiedName,
		))
	}
}

// checkFetchInHandler warns when fetch.Get, fetch.Post, or fetch.New is called
// directly inside an HTTP handler without SSRF protection.
func (a *Analyzer) checkFetchInHandler(qualifiedName string, expr *ast.MethodCallExpr) {
	fetchFunctions := map[string]bool{
		"fetch.Get":  true,
		"fetch.Post": true,
		"fetch.New":  true,
	}
	if !fetchFunctions[qualifiedName] {
		return
	}
	if !a.isInHTTPHandler() {
		return
	}
	a.error(expr.Pos(), fmt.Sprintf(
		"SSRF risk: %s inside an HTTP handler — use fetch.SafeGet or add fetch.Transport(netguard.HTTPTransport(...)) to restrict outbound requests",
		qualifiedName,
	))
}

// checkFilesInHandler warns when files.* I/O functions are called inside an
// HTTP handler, where the path argument may be user-controlled.
func (a *Analyzer) checkFilesInHandler(qualifiedName string, expr *ast.MethodCallExpr) {
	filesFunctions := map[string]bool{
		"files.Read":          true,
		"files.ReadBytes":     true,
		"files.Write":         true,
		"files.WriteString":   true,
		"files.Append":        true,
		"files.AppendString":  true,
		"files.Delete":        true,
		"files.DeleteAll":     true,
		"files.List":          true,
		"files.ListRecursive": true,
	}
	if !filesFunctions[qualifiedName] {
		return
	}
	if !a.isInHTTPHandler() {
		return
	}
	a.error(expr.Pos(), fmt.Sprintf(
		"path traversal risk: %s inside an HTTP handler — use sandbox.* with a restricted root for user-controlled paths",
		qualifiedName,
	))
}

// checkShellRunNonLiteral warns when shell.Run is called with a non-literal
// argument. shell.Run splits its argument on whitespace without quoting
// awareness; a variable value can silently inject extra arguments.
func (a *Analyzer) checkShellRunNonLiteral(qualifiedName string, expr *ast.MethodCallExpr, pipedArg *TypeInfo) {
	if qualifiedName != "shell.Run" {
		return
	}
	// Direct call: shell.Run(cmd) — cmd is at index 0.
	// Piped call: cmd |> shell.Run() — cmd is the piped value; skip since we
	// cannot verify a piped value's origin from TypeInfo alone.
	if pipedArg != nil {
		return
	}
	if len(expr.Arguments) == 0 {
		return
	}
	cmdArg := expr.Arguments[0]
	if _, ok := cmdArg.(*ast.StringLiteral); !ok {
		a.error(expr.Pos(), fmt.Sprintf(
			"command injection risk: shell.Run with non-literal argument — shell.Run splits on whitespace without quoting; use shell.Output() with separate arguments for variable input",
		))
	}
}

// checkRedirectNonLiteral warns when http.Redirect / http.RedirectPermanent is
// called with a non-literal URL argument, which is an open-redirect vector.
func (a *Analyzer) checkRedirectNonLiteral(qualifiedName string, expr *ast.MethodCallExpr, pipedArg *TypeInfo) {
	redirectFunctions := map[string]bool{
		"httphelper.Redirect":          true,
		"http.Redirect":                true,
		"httphelper.RedirectPermanent": true,
		"http.RedirectPermanent":       true,
	}
	if !redirectFunctions[qualifiedName] {
		return
	}
	// Stdlib files (e.g. http.kuki itself) are exempt: SafeRedirect and the
	// Redirect/RedirectPermanent wrappers call http.Redirect internally after
	// validation, so flagging them would produce false positives.
	if strings.Contains(a.sourceFile, "stdlib/") {
		return
	}
	// Redirect(w, r, url) — URL is the 3rd arg (index 2) in a plain call.
	// If one arg is piped (e.g. w |> Redirect(r, url)), URL is at index 1.
	urlArgIndex := 2
	if pipedArg != nil {
		urlArgIndex = 1
	}
	if urlArgIndex >= len(expr.Arguments) {
		return
	}
	urlArg := expr.Arguments[urlArgIndex]
	if _, ok := urlArg.(*ast.StringLiteral); !ok {
		a.error(expr.Pos(), fmt.Sprintf(
			"open redirect risk: %s with non-literal URL — use http.SafeRedirect(w, r, url, allowedHosts...) to validate the destination",
			qualifiedName,
		))
	}
}
