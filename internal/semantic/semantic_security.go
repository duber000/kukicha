package semantic

import (
	"fmt"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
)

// securityCategory returns the security check category for a qualified function
// name, checking both the generated registry and known aliases (e.g., httphelper.X → http.X).
func securityCategory(qualifiedName string) string {
	if cat := GetSecurityCategory(qualifiedName); cat != "" {
		return cat
	}
	// Handle aliases: httphelper.X → http.X
	if strings.HasPrefix(qualifiedName, "httphelper.") {
		suffix := qualifiedName[len("httphelper."):]
		return GetSecurityCategory("http." + suffix)
	}
	return ""
}

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
	if securityCategory(qualifiedName) != "sql" {
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
	if securityCategory(qualifiedName) != "html" {
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
	if securityCategory(qualifiedName) != "fetch" {
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
	if securityCategory(qualifiedName) != "files" {
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
	if securityCategory(qualifiedName) != "shell" {
		return
	}
	// Direct call: shell.Run(cmd) — cmd is at index 0.
	// Piped call: cmd |> shell.Run() — cmd is the piped value.
	if pipedArg != nil {
		// We can't verify the piped value's origin from TypeInfo alone,
		// but piping a variable into shell.Run is almost certainly unsafe.
		if pipedArg.Kind != TypeKindUnknown {
			a.warn(expr.Pos(),
				"command injection risk: piped value into shell.Run cannot be verified as safe — use shell.Output() with separate arguments for variable input")
		}
		return
	}
	if len(expr.Arguments) == 0 {
		return
	}
	cmdArg := expr.Arguments[0]
	if _, ok := cmdArg.(*ast.StringLiteral); !ok {
		a.error(expr.Pos(),
			"command injection risk: shell.Run with non-literal argument — shell.Run splits on whitespace without quoting; use shell.Output() with separate arguments for variable input",
		)
	}
}

// checkRedirectNonLiteral warns when http.Redirect / http.RedirectPermanent is
// called with a non-literal URL argument, which is an open-redirect vector.
func (a *Analyzer) checkRedirectNonLiteral(qualifiedName string, expr *ast.MethodCallExpr, pipedArg *TypeInfo) {
	if securityCategory(qualifiedName) != "redirect" {
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
