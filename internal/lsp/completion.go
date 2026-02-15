package lsp

import (
	"context"
	"encoding/json"
	"log"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// handleCompletion handles textDocument/completion requests
func (s *Server) handleCompletion(ctx context.Context, req *jsonrpc2.Request) (*lsp.CompletionList, error) {
	var params lsp.CompletionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	doc := s.documents.Get(params.TextDocument.URI)
	if doc == nil {
		return &lsp.CompletionList{}, nil
	}

	log.Printf("Completion request at %d:%d", params.Position.Line, params.Position.Character)

	items := s.getCompletions(doc, params.Position)

	return &lsp.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}, nil
}

// handleDocumentSymbol handles textDocument/documentSymbol requests
func (s *Server) handleDocumentSymbol(ctx context.Context, req *jsonrpc2.Request) ([]lsp.SymbolInformation, error) {
	var params lsp.DocumentSymbolParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	doc := s.documents.Get(params.TextDocument.URI)
	if doc == nil {
		return nil, nil
	}

	return s.getDocumentSymbols(doc), nil
}

// getCompletions returns completion items for the given position
func (s *Server) getCompletions(doc *Document, pos lsp.Position) []lsp.CompletionItem {
	items := []lsp.CompletionItem{}

	// Add keywords
	keywords := []string{
		"func", "function", "type", "interface", "petiole", "import",
		"if", "else", "for", "in", "from", "to", "through",
		"switch", "when", "otherwise", "default",
		"return", "break", "continue", "defer", "go",
		"true", "false", "empty", "nil", "make", "onerr",
		"and", "or", "not", "equals", "reference", "dereference",
		"send", "receive", "many", "channel", "list", "map", "of", "as",
		"variable", "var", "on", "close", "panic", "error", "discard",
	}
	for _, kw := range keywords {
		items = append(items, lsp.CompletionItem{
			Label:  kw,
			Kind:   lsp.CIKKeyword,
			Detail: "keyword",
		})
	}

	// Add builtin functions
	builtins := []struct {
		name   string
		detail string
	}{
		{"print", "func print(args ...any)"},
		{"len", "func len(v any) int"},
		{"append", "func append(slice []T, elems ...T) []T"},
		{"make", "func make(T type, size ...int) T"},
		{"close", "func close(ch chan T)"},
		{"panic", "func panic(v any)"},
		{"recover", "func recover() any"},
	}
	for _, b := range builtins {
		items = append(items, lsp.CompletionItem{
			Label:  b.name,
			Kind:   lsp.CIKFunction,
			Detail: b.detail,
		})
	}

	// Add primitive types
	types := []string{
		"int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64",
		"float32", "float64", "string", "bool", "byte", "rune", "any", "error",
	}
	for _, t := range types {
		items = append(items, lsp.CompletionItem{
			Label:  t,
			Kind:   lsp.CIKTypeParameter,
			Detail: "type",
		})
	}

	// Add declarations from the current document
	if doc.Program != nil {
		for _, decl := range doc.Program.Declarations {
			switch d := decl.(type) {
			case *ast.FunctionDecl:
				items = append(items, lsp.CompletionItem{
					Label:  d.Name.Value,
					Kind:   lsp.CIKFunction,
					Detail: formatFunctionDecl(d),
				})
			case *ast.TypeDecl:
				items = append(items, lsp.CompletionItem{
					Label:  d.Name.Value,
					Kind:   lsp.CIKStruct,
					Detail: "type",
				})
			case *ast.InterfaceDecl:
				items = append(items, lsp.CompletionItem{
					Label:  d.Name.Value,
					Kind:   lsp.CIKInterface,
					Detail: "interface",
				})
			}
		}
	}

	return items
}

// getDocumentSymbols returns all symbols in the document
func (s *Server) getDocumentSymbols(doc *Document) []lsp.SymbolInformation {
	symbols := []lsp.SymbolInformation{}

	if doc.Program == nil {
		return symbols
	}

	// Add petiole declaration
	if doc.Program.PetioleDecl != nil {
		symbols = append(symbols, lsp.SymbolInformation{
			Name: doc.Program.PetioleDecl.Name.Value,
			Kind: lsp.SKPackage,
			Location: lsp.Location{
				URI: doc.URI,
				Range: lsp.Range{
					Start: lsp.Position{
						Line:      doc.Program.PetioleDecl.Pos().Line - 1,
						Character: doc.Program.PetioleDecl.Pos().Column - 1,
					},
					End: lsp.Position{
						Line:      doc.Program.PetioleDecl.Pos().Line - 1,
						Character: doc.Program.PetioleDecl.Pos().Column - 1 + len(doc.Program.PetioleDecl.Name.Value),
					},
				},
			},
		})
	}

	// Add top-level declarations
	for _, decl := range doc.Program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			kind := lsp.SKFunction
			if d.Receiver != nil {
				kind = lsp.SKMethod
			}
			symbols = append(symbols, lsp.SymbolInformation{
				Name: d.Name.Value,
				Kind: kind,
				Location: lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      d.Pos().Line - 1,
							Character: d.Pos().Column - 1,
						},
						End: lsp.Position{
							Line:      d.Pos().Line - 1,
							Character: d.Pos().Column - 1 + len(d.Name.Value),
						},
					},
				},
			})
		case *ast.TypeDecl:
			symbols = append(symbols, lsp.SymbolInformation{
				Name: d.Name.Value,
				Kind: lsp.SKStruct,
				Location: lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      d.Pos().Line - 1,
							Character: d.Pos().Column - 1,
						},
						End: lsp.Position{
							Line:      d.Pos().Line - 1,
							Character: d.Pos().Column - 1 + len(d.Name.Value),
						},
					},
				},
			})
			// Add fields
			for _, field := range d.Fields {
				symbols = append(symbols, lsp.SymbolInformation{
					Name:          field.Name.Value,
					Kind:          lsp.SKField,
					ContainerName: d.Name.Value,
					Location: lsp.Location{
						URI: doc.URI,
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      field.Name.Pos().Line - 1,
								Character: field.Name.Pos().Column - 1,
							},
							End: lsp.Position{
								Line:      field.Name.Pos().Line - 1,
								Character: field.Name.Pos().Column - 1 + len(field.Name.Value),
							},
						},
					},
				})
			}
		case *ast.InterfaceDecl:
			symbols = append(symbols, lsp.SymbolInformation{
				Name: d.Name.Value,
				Kind: lsp.SKInterface,
				Location: lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      d.Pos().Line - 1,
							Character: d.Pos().Column - 1,
						},
						End: lsp.Position{
							Line:      d.Pos().Line - 1,
							Character: d.Pos().Column - 1 + len(d.Name.Value),
						},
					},
				},
			})
			// Add interface methods
			for _, method := range d.Methods {
				symbols = append(symbols, lsp.SymbolInformation{
					Name:          method.Name.Value,
					Kind:          lsp.SKMethod,
					ContainerName: d.Name.Value,
					Location: lsp.Location{
						URI: doc.URI,
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      method.Name.Pos().Line - 1,
								Character: method.Name.Pos().Column - 1,
							},
							End: lsp.Position{
								Line:      method.Name.Pos().Line - 1,
								Character: method.Name.Pos().Column - 1 + len(method.Name.Value),
							},
						},
					},
				})
			}
		}
	}

	return symbols
}
