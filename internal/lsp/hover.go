package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/semantic"
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

	// Look for variables/parameters - would need proper scope analysis
	// For now, return type information if available
	return ""
}

// getBuiltinInfo returns documentation for builtin functions
func getBuiltinInfo(name string) string {
	builtins := map[string]string{
		"print":  "func print(args ...any)\nPrints values to stdout",
		"len":    "func len(v any) int\nReturns the length of a string, list, or map",
		"append": "func append(slice []T, elems ...T) []T\nAppends elements to a slice",
		"make":   "func make(T type, size ...int) T\nCreates a slice, map, or channel",
		"close":  "func close(ch chan T)\nCloses a channel",
		"panic":  "func panic(v any)\nStops normal execution and begins panicking",
		"recover": "func recover() any\nRegains control of a panicking goroutine",
		"empty":  "empty T\nReturns the zero value of type T",
		"error":  "error \"message\"\nCreates a new error with the given message",
	}

	if info, ok := builtins[name]; ok {
		return info
	}
	return ""
}

// formatFunctionDecl formats a function declaration for hover display
func formatFunctionDecl(decl *ast.FunctionDecl) string {
	var result string

	// Add receiver if it's a method
	if decl.Receiver != nil {
		result += fmt.Sprintf("func (%s %s) ", decl.Receiver.Name.Value, formatTypeAnnotation(decl.Receiver.Type))
	} else {
		result += "func "
	}

	result += decl.Name.Value + "("

	// Parameters
	for i, param := range decl.Parameters {
		if i > 0 {
			result += ", "
		}
		if param.Variadic {
			result += "many "
		}
		result += param.Name.Value + " " + formatTypeAnnotation(param.Type)
	}
	result += ")"

	// Returns
	if len(decl.Returns) > 0 {
		if len(decl.Returns) == 1 {
			result += " " + formatTypeAnnotation(decl.Returns[0])
		} else {
			result += " ("
			for i, ret := range decl.Returns {
				if i > 0 {
					result += ", "
				}
				result += formatTypeAnnotation(ret)
			}
			result += ")"
		}
	}

	return result
}

// formatTypeDecl formats a type declaration for hover display
func formatTypeDecl(decl *ast.TypeDecl) string {
	result := fmt.Sprintf("type %s\n", decl.Name.Value)

	if len(decl.Fields) > 0 {
		result += "Fields:\n"
		for _, field := range decl.Fields {
			result += fmt.Sprintf("  %s %s\n", field.Name.Value, formatTypeAnnotation(field.Type))
		}
	}

	return result
}

// formatInterfaceDecl formats an interface declaration for hover display
func formatInterfaceDecl(decl *ast.InterfaceDecl) string {
	result := fmt.Sprintf("interface %s\n", decl.Name.Value)

	if len(decl.Methods) > 0 {
		result += "Methods:\n"
		for _, method := range decl.Methods {
			result += fmt.Sprintf("  %s(", method.Name.Value)
			for i, param := range method.Parameters {
				if i > 0 {
					result += ", "
				}
				result += param.Name.Value + " " + formatTypeAnnotation(param.Type)
			}
			result += ")"
			if len(method.Returns) > 0 {
				result += " "
				for i, ret := range method.Returns {
					if i > 0 {
						result += ", "
					}
					result += formatTypeAnnotation(ret)
				}
			}
			result += "\n"
		}
	}

	return result
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
		result := "func("
		for i, param := range ta.Parameters {
			if i > 0 {
				result += ", "
			}
			result += formatTypeAnnotation(param)
		}
		result += ")"
		if len(ta.Returns) > 0 {
			result += " "
			for i, ret := range ta.Returns {
				if i > 0 {
					result += ", "
				}
				result += formatTypeAnnotation(ret)
			}
		}
		return result
	default:
		return "unknown"
	}
}

// findSymbolInScope searches for a symbol definition
func findSymbolInScope(program *ast.Program, name string) *semantic.Symbol {
	// This is a simplified version - a full implementation would need
	// to track the cursor position and find the appropriate scope
	return nil
}
