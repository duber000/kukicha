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
		a.errorDiag(strLit.Pos(),
			fmt.Sprintf("SQL injection risk: string interpolation in %s query — use parameter placeholders ($1, $2, ...) instead", qualifiedName),
			"security/sql-injection",
			"use parameter placeholders ($1, $2, ...) instead of string interpolation",
		)
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
		a.errorDiag(expr.Pos(),
			fmt.Sprintf("XSS risk: %s with non-literal content — use http.SafeHTML to HTML-escape user-controlled content", qualifiedName),
			"security/xss",
			"use http.SafeHTML to HTML-escape user-controlled content, or use http.Text() for plain text",
		)
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
	a.errorDiag(expr.Pos(),
		fmt.Sprintf("SSRF risk: %s inside an HTTP handler — use fetch.SafeGet or add fetch.Transport(netguard.HTTPTransport(...)) to restrict outbound requests", qualifiedName),
		"security/ssrf",
		"use fetch.SafeGet or add fetch.Transport(netguard.HTTPTransport(...)) to restrict outbound requests",
	)
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
	a.errorDiag(expr.Pos(),
		fmt.Sprintf("path traversal risk: %s inside an HTTP handler — use sandbox.* with a restricted root for user-controlled paths", qualifiedName),
		"security/path-traversal",
		"use sandbox.* with a restricted root for user-controlled paths",
	)
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
			a.warnDiag(expr.Pos(),
				"command injection risk: piped value into shell.Run cannot be verified as safe — use shell.Output() with separate arguments for variable input",
				"security/command-injection",
				"use shell.Output() with separate arguments for variable input",
			)
		}
		return
	}
	if len(expr.Arguments) == 0 {
		return
	}
	cmdArg := expr.Arguments[0]
	if _, ok := cmdArg.(*ast.StringLiteral); !ok {
		a.errorDiag(expr.Pos(),
			"command injection risk: shell.Run with non-literal argument — shell.Run splits on whitespace without quoting; use shell.Output() with separate arguments for variable input",
			"security/command-injection",
			"use shell.Output() with separate arguments for variable input",
		)
	}
}

// --- Tier 5: Agent-specific security checks ---

// checkUnboundedLoop warns when a for loop with no obvious termination condition
// appears inside an HTTP handler. Infinite loops in handlers cause goroutine leaks.
func (a *Analyzer) checkUnboundedLoop(stmt *ast.ForConditionStmt) {
	if !a.isInHTTPHandler() {
		return
	}
	// Check if the condition is a literal true (e.g., "for true")
	if boolLit, ok := stmt.Condition.(*ast.BooleanLiteral); ok && boolLit.Value {
		// Check if the body contains break/return — if so, it's bounded
		if !blockContainsTerminator(stmt.Body) {
			a.warnDiag(stmt.Pos(),
				"unbounded loop: 'for true' inside an HTTP handler with no break/return — this may cause goroutine leaks",
				"agent/unbounded-loop",
				"add a break condition, context deadline, or iteration limit to prevent unbounded execution",
			)
		}
	}
}

// blockContainsTerminator returns true if a block contains a break, return,
// or continue statement at any nesting depth.
func blockContainsTerminator(block *ast.BlockStmt) bool {
	if block == nil {
		return false
	}
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.BreakStmt, *ast.ContinueStmt, *ast.ReturnStmt:
			return true
		case *ast.IfStmt:
			if blockContainsTerminator(s.Consequence) {
				return true
			}
			switch alt := s.Alternative.(type) {
			case *ast.ElseStmt:
				if blockContainsTerminator(alt.Body) {
					return true
				}
			}
		}
	}
	return false
}

// checkResourceExhaustion warns about potential resource exhaustion patterns
// in HTTP handlers: unbounded goroutine spawning or channel creation inside loops.
func (a *Analyzer) checkResourceExhaustion(stmt ast.Statement) {
	if !a.isInHTTPHandler() {
		return
	}
	if a.loopDepth == 0 {
		return
	}

	switch s := stmt.(type) {
	case *ast.GoStmt:
		a.warnDiag(s.Pos(),
			"resource exhaustion risk: goroutine spawned inside a loop in an HTTP handler — unbounded goroutine creation can exhaust memory",
			"agent/resource-exhaustion",
			"use a worker pool or limit concurrency with a semaphore channel",
		)
	case *ast.VarDeclStmt:
		// Check for channel creation inside loops (make channel of T)
		for _, val := range s.Values {
			if makeExpr, ok := val.(*ast.MakeExpr); ok {
				if _, ok := makeExpr.Type.(*ast.ChannelType); ok {
					a.warnDiag(s.Pos(),
						"resource exhaustion risk: channel created inside a loop in an HTTP handler — unbounded channel allocation can exhaust memory",
						"agent/resource-exhaustion",
						"create channels outside the loop or limit loop iterations",
					)
				}
			}
		}
	}
}

// checkPrivilegeEscalation warns when shell.Run or shell.Command is called
// with arguments derived from HTTP request parameters inside a handler.
// This extends checkShellRunNonLiteral with taint tracking from HTTP params.
func (a *Analyzer) checkPrivilegeEscalation(qualifiedName string, expr *ast.MethodCallExpr, pipedArg *TypeInfo) {
	if securityCategory(qualifiedName) != "shell" {
		return
	}
	if !a.isInHTTPHandler() {
		return
	}
	// Already covered by checkShellRunNonLiteral for non-literal args.
	// This adds the handler context warning for any shell usage in handlers.
	a.warnDiag(expr.Pos(),
		fmt.Sprintf("privilege escalation risk: %s inside an HTTP handler — shell commands in handlers may execute with server privileges on user-controlled input", qualifiedName),
		"agent/privilege-escalation",
		"avoid shell commands in HTTP handlers, or validate and sanitize all inputs before passing to shell",
	)
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
		a.errorDiag(expr.Pos(),
			fmt.Sprintf("open redirect risk: %s with non-literal URL — use http.SafeRedirect(w, r, url, allowedHosts...) to validate the destination", qualifiedName),
			"security/open-redirect",
			"use http.SafeRedirect(w, r, url, allowedHosts...) to validate the destination",
		)
	}
}
