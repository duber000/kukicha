package lsp

import (
	"context"
	"log"
	"regexp"
	"strconv"

	"github.com/sourcegraph/go-lsp"
)

// publishDiagnostics analyzes the document and publishes diagnostics to the client
func (s *Server) publishDiagnostics(ctx context.Context, uri lsp.DocumentURI) {
	doc := s.documents.Get(uri)
	if doc == nil {
		return
	}

	diagnostics := make([]lsp.Diagnostic, 0, len(doc.Errors))

	for _, err := range doc.Errors {
		diag := errorToDiagnostic(err)
		diagnostics = append(diagnostics, diag)
	}

	log.Printf("Publishing %d diagnostics for %s", len(diagnostics), uri)

	s.conn.Notify(ctx, "textDocument/publishDiagnostics", &lsp.PublishDiagnosticsParams{
		URI:         uri,
		Diagnostics: diagnostics,
	})
}

// errorToDiagnostic converts a compiler error to an LSP diagnostic
func errorToDiagnostic(err error) lsp.Diagnostic {
	msg := err.Error()

	// Parse error format: "filename:line:column: message"
	// Example: "test.kuki:5:10: undefined identifier 'foo'"
	re := regexp.MustCompile(`^(.+):(\d+):(\d+): (.+)$`)
	matches := re.FindStringSubmatch(msg)

	var line, col int
	var message string

	if len(matches) == 5 {
		line, _ = strconv.Atoi(matches[2])
		col, _ = strconv.Atoi(matches[3])
		message = matches[4]
		// Convert to 0-indexed
		line--
		col--
	} else {
		// Fallback: put error at beginning of file
		line = 0
		col = 0
		message = msg
	}

	// Ensure non-negative values
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}

	return lsp.Diagnostic{
		Range: lsp.Range{
			Start: lsp.Position{Line: line, Character: col},
			End:   lsp.Position{Line: line, Character: col + 1},
		},
		Severity: lsp.Error,
		Source:   "kukicha",
		Message:  message,
	}
}
