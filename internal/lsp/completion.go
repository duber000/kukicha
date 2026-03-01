package lsp

import (
	"context"
	"encoding/json"
	"log"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/lexer"
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

	// Add keywords (from the lexer's canonical list)
	for _, kw := range lexer.Keywords() {
		items = append(items, lsp.CompletionItem{
			Label:  kw,
			Kind:   lsp.CIKKeyword,
			Detail: "keyword",
		})
	}

	// Add builtin functions (from the shared registry)
	for _, b := range builtinCompletions() {
		items = append(items, lsp.CompletionItem{
			Label:  b.Name,
			Kind:   lsp.CIKFunction,
			Detail: b.Signature,
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

	// Helper to ensure non-negative positions
	pos := func(line, col int) lsp.Position {
		return lsp.Position{
			Line:      max(0, line-1),
			Character: max(0, col-1),
		}
	}

	// Add petiole declaration
	if doc.Program.PetioleDecl != nil {
		startCol := doc.Program.PetioleDecl.Pos().Column - 1
		endCol := startCol + len(doc.Program.PetioleDecl.Name.Value)
		symbols = append(symbols, lsp.SymbolInformation{
			Name: doc.Program.PetioleDecl.Name.Value,
			Kind: lsp.SKPackage,
			Location: lsp.Location{
				URI: doc.URI,
				Range: lsp.Range{
					Start: pos(doc.Program.PetioleDecl.Pos().Line, startCol+1),
					End:   pos(doc.Program.PetioleDecl.Pos().Line, endCol+1),
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
			startCol := d.Pos().Column - 1
			endCol := startCol + len(d.Name.Value)
			symbols = append(symbols, lsp.SymbolInformation{
				Name: d.Name.Value,
				Kind: kind,
				Location: lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: pos(d.Pos().Line, startCol+1),
						End:   pos(d.Pos().Line, endCol+1),
					},
				},
			})
		case *ast.TypeDecl:
			startCol := d.Pos().Column - 1
			endCol := startCol + len(d.Name.Value)
			symbols = append(symbols, lsp.SymbolInformation{
				Name: d.Name.Value,
				Kind: lsp.SKStruct,
				Location: lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: pos(d.Pos().Line, startCol+1),
						End:   pos(d.Pos().Line, endCol+1),
					},
				},
			})
			// Add fields
			for _, field := range d.Fields {
				fStartCol := field.Name.Pos().Column - 1
				fEndCol := fStartCol + len(field.Name.Value)
				symbols = append(symbols, lsp.SymbolInformation{
					Name:          field.Name.Value,
					Kind:          lsp.SKField,
					ContainerName: d.Name.Value,
					Location: lsp.Location{
						URI: doc.URI,
						Range: lsp.Range{
							Start: pos(field.Name.Pos().Line, fStartCol+1),
							End:   pos(field.Name.Pos().Line, fEndCol+1),
						},
					},
				})
			}
		case *ast.InterfaceDecl:
			startCol := d.Pos().Column - 1
			endCol := startCol + len(d.Name.Value)
			symbols = append(symbols, lsp.SymbolInformation{
				Name: d.Name.Value,
				Kind: lsp.SKInterface,
				Location: lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: pos(d.Pos().Line, startCol+1),
						End:   pos(d.Pos().Line, endCol+1),
					},
				},
			})
			// Add interface methods
			for _, method := range d.Methods {
				mStartCol := method.Name.Pos().Column - 1
				mEndCol := mStartCol + len(method.Name.Value)
				symbols = append(symbols, lsp.SymbolInformation{
					Name:          method.Name.Value,
					Kind:          lsp.SKMethod,
					ContainerName: d.Name.Value,
					Location: lsp.Location{
						URI: doc.URI,
						Range: lsp.Range{
							Start: pos(method.Name.Pos().Line, mStartCol+1),
							End:   pos(method.Name.Pos().Line, mEndCol+1),
						},
					},
				})
			}
		}
	}

	return symbols
}
