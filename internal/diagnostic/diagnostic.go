// Package diagnostic defines structured compiler diagnostics for Kukicha.
// Diagnostics carry machine-readable metadata (file, line, col, category,
// suggestion) suitable for JSON output and AI agent consumption.
//
// Diagnostic implements the error interface so it can be stored transparently
// in []error slices, maintaining backward compatibility with existing code.
package diagnostic

import "fmt"

// Diagnostic represents a structured compiler diagnostic with metadata
// for programmatic consumption (JSON output, AI agent loops).
type Diagnostic struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
	Severity   string `json:"severity"`              // "error" or "warning"
	Category   string `json:"category,omitempty"`     // e.g. "security/sql-injection"
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

// Error implements the error interface, producing the same format as the
// existing plain-text error output: "file:line:col: message".
func (d Diagnostic) Error() string {
	return fmt.Sprintf("%s:%d:%d: %s", d.File, d.Line, d.Col, d.Message)
}
