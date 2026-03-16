package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// handleHover handles textDocument/hover requests
func (s *Server) handleHover(ctx context.Context, req *jsonrpc2.Request) (*lsp.Hover, error) {
	var params lsp.TextDocumentPositionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	doc := s.documents.Get(params.TextDocument.URI)
	if doc == nil {
		return nil, nil
	}

	// Get the word at the cursor position
	word := doc.GetWordAtPosition(params.Position)
	if word == "" {
		return nil, nil
	}

	log.Printf("Hover request for word: %s at %d:%d", word, params.Position.Line, params.Position.Character)

	// Look up the symbol in the program
	hoverContent := s.getHoverContent(doc, word, params.Position)
	if hoverContent == "" {
		return nil, nil
	}

	return &lsp.Hover{
		Contents: []lsp.MarkedString{
			{Language: "kukicha", Value: hoverContent},
		},
		Range: &lsp.Range{
			Start: params.Position,
			End: lsp.Position{
				Line:      params.Position.Line,
				Character: params.Position.Character + len(word),
			},
		},
	}, nil
}

// getHoverContent returns hover information for a symbol
func (s *Server) getHoverContent(doc *Document, word string, pos lsp.Position) string {
	if doc.Program == nil {
		return ""
	}

	// Check builtins first
	if builtin := getBuiltinInfo(word); builtin != "" {
		return builtin
	}

	// Search for declarations
	for _, decl := range doc.Program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			if d.Name.Value == word {
				return formatFunctionDecl(d)
			}
		case *ast.TypeDecl:
			if d.Name.Value == word {
				return formatTypeDecl(d)
			}
		case *ast.InterfaceDecl:
			if d.Name.Value == word {
				return formatInterfaceDecl(d)
			}
		}
	}

	// Look for local variables and parameters inside the function at the cursor position
	if result := findLocalSymbol(doc.Program, word, pos); result != "" {
		return result
	}

	return ""
}

// getBuiltinInfo returns documentation for builtin functions
// using the shared builtin registry in builtins.go.
func getBuiltinInfo(name string) string {
	return lookupBuiltin(name)
}

// formatFunctionDecl formats a function declaration for hover display
func formatFunctionDecl(decl *ast.FunctionDecl) string {
	var result strings.Builder

	// Add receiver if it's a method
	if decl.Receiver != nil {
		result.WriteString(fmt.Sprintf("func (%s %s) ", decl.Receiver.Name.Value, formatTypeAnnotation(decl.Receiver.Type)))
	} else {
		result.WriteString("func ")
	}

	result.WriteString(decl.Name.Value + "(")

	// Parameters
	for i, param := range decl.Parameters {
		if i > 0 {
			result.WriteString(", ")
		}
		if param.Variadic {
			result.WriteString("many ")
		}
		result.WriteString(param.Name.Value + " " + formatTypeAnnotation(param.Type))
	}
	result.WriteString(")")

	// Returns
	if len(decl.Returns) > 0 {
		if len(decl.Returns) == 1 {
			result.WriteString(" " + formatTypeAnnotation(decl.Returns[0]))
		} else {
			result.WriteString(" (")
			for i, ret := range decl.Returns {
				if i > 0 {
					result.WriteString(", ")
				}
				result.WriteString(formatTypeAnnotation(ret))
			}
			result.WriteString(")")
		}
	}

	return result.String()
}

// formatTypeDecl formats a type declaration for hover display
func formatTypeDecl(decl *ast.TypeDecl) string {
	if decl.AliasType != nil {
		return fmt.Sprintf("type %s %s", decl.Name.Value, formatTypeAnnotation(decl.AliasType))
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("type %s\n", decl.Name.Value))

	if len(decl.Fields) > 0 {
		result.WriteString("Fields:\n")
		for _, field := range decl.Fields {
			result.WriteString(fmt.Sprintf("  %s %s\n", field.Name.Value, formatTypeAnnotation(field.Type)))
		}
	}

	return result.String()
}

// formatInterfaceDecl formats an interface declaration for hover display
func formatInterfaceDecl(decl *ast.InterfaceDecl) string {
	var result strings.Builder
	result.WriteString(fmt.Sprintf("interface %s\n", decl.Name.Value))

	if len(decl.Methods) > 0 {
		result.WriteString("Methods:\n")
		for _, method := range decl.Methods {
			result.WriteString(fmt.Sprintf("  %s(", method.Name.Value))
			for i, param := range method.Parameters {
				if i > 0 {
					result.WriteString(", ")
				}
				result.WriteString(param.Name.Value + " " + formatTypeAnnotation(param.Type))
			}
			result.WriteString(")")
			if len(method.Returns) > 0 {
				result.WriteString(" ")
				for i, ret := range method.Returns {
					if i > 0 {
						result.WriteString(", ")
					}
					result.WriteString(formatTypeAnnotation(ret))
				}
			}
			result.WriteString("\n")
		}
	}

	return result.String()
}

// formatTypeAnnotation converts a type annotation to a string
func formatTypeAnnotation(t ast.TypeAnnotation) string {
	if t == nil {
		return "unknown"
	}

	switch ta := t.(type) {
	case *ast.PrimitiveType:
		return ta.Name
	case *ast.NamedType:
		return ta.Name
	case *ast.ReferenceType:
		return "reference " + formatTypeAnnotation(ta.ElementType)
	case *ast.ListType:
		return "list of " + formatTypeAnnotation(ta.ElementType)
	case *ast.MapType:
		return fmt.Sprintf("map of %s to %s", formatTypeAnnotation(ta.KeyType), formatTypeAnnotation(ta.ValueType))
	case *ast.ChannelType:
		return "channel of " + formatTypeAnnotation(ta.ElementType)
	case *ast.FunctionType:
		var result strings.Builder
		result.WriteString("func(")
		for i, param := range ta.Parameters {
			if i > 0 {
				result.WriteString(", ")
			}
			result.WriteString(formatTypeAnnotation(param))
		}
		result.WriteString(")")
		if len(ta.Returns) > 0 {
			result.WriteString(" ")
			for i, ret := range ta.Returns {
				if i > 0 {
					result.WriteString(", ")
				}
				result.WriteString(formatTypeAnnotation(ret))
			}
		}
		return result.String()
	default:
		return "unknown"
	}
}

// findLocalSymbol searches function bodies for a local variable or parameter
// matching the given word at the cursor position. It walks the AST to find
// the enclosing function, then checks parameters and variable declarations.
func findLocalSymbol(program *ast.Program, word string, pos lsp.Position) string {
	// LSP positions are 0-indexed; AST positions are 1-indexed
	cursorLine := int(pos.Line) + 1

	for _, decl := range program.Declarations {
		fn, ok := decl.(*ast.FunctionDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// Check if cursor is within this function's body
		if !blockContainsLine(fn.Body, cursorLine) {
			continue
		}

		// Check receiver
		if fn.Receiver != nil && fn.Receiver.Name.Value == word {
			return fmt.Sprintf("%s %s (receiver)", fn.Receiver.Name.Value, formatTypeAnnotation(fn.Receiver.Type))
		}

		// Check parameters
		for _, param := range fn.Parameters {
			if param.Name.Value == word {
				prefix := ""
				if param.Variadic {
					prefix = "many "
				}
				return fmt.Sprintf("%s%s %s (parameter)", prefix, param.Name.Value, formatTypeAnnotation(param.Type))
			}
		}

		// Search block for variable declarations
		if result := findVarInBlock(fn.Body, word, cursorLine); result != "" {
			return result
		}

		// Found enclosing function but no match
		return ""
	}
	return ""
}

// blockContainsLine returns true if the block's line range contains the given line.
func blockContainsLine(block *ast.BlockStmt, line int) bool {
	if block == nil || len(block.Statements) == 0 {
		return false
	}
	startLine := block.Pos().Line
	endLine := lastLineInBlock(block)
	return line >= startLine && line <= endLine
}

// lastLineInBlock estimates the last line of a block by checking its statements.
func lastLineInBlock(block *ast.BlockStmt) int {
	if block == nil || len(block.Statements) == 0 {
		return 0
	}
	last := block.Statements[len(block.Statements)-1]
	line := last.Pos().Line

	// For nested blocks, recurse to find the actual last line
	switch s := last.(type) {
	case *ast.IfStmt:
		if alt := s.Alternative; alt != nil {
			switch a := alt.(type) {
			case *ast.ElseStmt:
				if end := lastLineInBlock(a.Body); end > line {
					line = end
				}
			case *ast.IfStmt:
				// else if — estimate via its consequence
				if end := lastLineInBlock(a.Consequence); end > line {
					line = end
				}
			}
		} else if s.Consequence != nil {
			if end := lastLineInBlock(s.Consequence); end > line {
				line = end
			}
		}
	case *ast.ForRangeStmt:
		if end := lastLineInBlock(s.Body); end > line {
			line = end
		}
	case *ast.ForNumericStmt:
		if end := lastLineInBlock(s.Body); end > line {
			line = end
		}
	case *ast.ForConditionStmt:
		if end := lastLineInBlock(s.Body); end > line {
			line = end
		}
	case *ast.SwitchStmt:
		for _, c := range s.Cases {
			if end := lastLineInBlock(c.Body); end > line {
				line = end
			}
		}
		if s.Otherwise != nil {
			if end := lastLineInBlock(s.Otherwise.Body); end > line {
				line = end
			}
		}
	}
	return line
}

// findVarInBlock searches a block for variable declarations matching word,
// only returning declarations that appear before the cursor line.
func findVarInBlock(block *ast.BlockStmt, word string, cursorLine int) string {
	if block == nil {
		return ""
	}
	for _, stmt := range block.Statements {
		// Only consider declarations before or at the cursor line
		if stmt.Pos().Line > cursorLine {
			break
		}

		switch s := stmt.(type) {
		case *ast.VarDeclStmt:
			for _, name := range s.Names {
				if name.Value == word {
					if s.Type != nil {
						return fmt.Sprintf("%s %s (variable)", word, formatTypeAnnotation(s.Type))
					}
					return fmt.Sprintf("%s (variable)", word)
				}
			}

		case *ast.ForRangeStmt:
			if blockContainsLine(s.Body, cursorLine) {
				// Check loop variable
				if s.Variable != nil && s.Variable.Value == word {
					return fmt.Sprintf("%s (range variable)", word)
				}
				if s.Index != nil && s.Index.Value == word {
					return fmt.Sprintf("%s (range index)", word)
				}
				// Recurse into body
				if result := findVarInBlock(s.Body, word, cursorLine); result != "" {
					return result
				}
			}

		case *ast.ForNumericStmt:
			if blockContainsLine(s.Body, cursorLine) {
				if s.Variable != nil && s.Variable.Value == word {
					return fmt.Sprintf("%s int (loop variable)", word)
				}
				if result := findVarInBlock(s.Body, word, cursorLine); result != "" {
					return result
				}
			}

		case *ast.ForConditionStmt:
			if blockContainsLine(s.Body, cursorLine) {
				if result := findVarInBlock(s.Body, word, cursorLine); result != "" {
					return result
				}
			}

		case *ast.IfStmt:
			if s.Consequence != nil && blockContainsLine(s.Consequence, cursorLine) {
				if result := findVarInBlock(s.Consequence, word, cursorLine); result != "" {
					return result
				}
			}
			if s.Alternative != nil {
				switch alt := s.Alternative.(type) {
				case *ast.ElseStmt:
					if alt.Body != nil && blockContainsLine(alt.Body, cursorLine) {
						if result := findVarInBlock(alt.Body, word, cursorLine); result != "" {
							return result
						}
					}
				case *ast.IfStmt:
					// Recurse for else-if chains
					if alt.Consequence != nil && blockContainsLine(alt.Consequence, cursorLine) {
						if result := findVarInBlock(alt.Consequence, word, cursorLine); result != "" {
							return result
						}
					}
				}
			}

		case *ast.SwitchStmt:
			for _, c := range s.Cases {
				if c.Body != nil && blockContainsLine(c.Body, cursorLine) {
					if result := findVarInBlock(c.Body, word, cursorLine); result != "" {
						return result
					}
				}
			}
			if s.Otherwise != nil && s.Otherwise.Body != nil && blockContainsLine(s.Otherwise.Body, cursorLine) {
				if result := findVarInBlock(s.Otherwise.Body, word, cursorLine); result != "" {
					return result
				}
			}
		}
	}
	return ""
}

