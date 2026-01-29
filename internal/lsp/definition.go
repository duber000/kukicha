package lsp

import (
	"context"
	"encoding/json"
	"log"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// handleDefinition handles textDocument/definition requests
func (s *Server) handleDefinition(ctx context.Context, req *jsonrpc2.Request) ([]lsp.Location, error) {
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

	log.Printf("Definition request for word: %s", word)

	// Find definition
	location := s.findDefinition(doc, word)
	if location == nil {
		return nil, nil
	}

	return []lsp.Location{*location}, nil
}

// findDefinition finds the definition location for a symbol
func (s *Server) findDefinition(doc *Document, word string) *lsp.Location {
	if doc.Program == nil {
		return nil
	}

	// Search top-level declarations
	for _, decl := range doc.Program.Declarations {
		switch d := decl.(type) {
		case *ast.FunctionDecl:
			if d.Name.Value == word {
				return &lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      d.Name.Pos().Line - 1, // Convert to 0-indexed
							Character: d.Name.Pos().Column - 1,
						},
						End: lsp.Position{
							Line:      d.Name.Pos().Line - 1,
							Character: d.Name.Pos().Column - 1 + len(word),
						},
					},
				}
			}
		case *ast.TypeDecl:
			if d.Name.Value == word {
				return &lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      d.Name.Pos().Line - 1,
							Character: d.Name.Pos().Column - 1,
						},
						End: lsp.Position{
							Line:      d.Name.Pos().Line - 1,
							Character: d.Name.Pos().Column - 1 + len(word),
						},
					},
				}
			}
			// Check for field definitions
			for _, field := range d.Fields {
				if field.Name.Value == word {
					return &lsp.Location{
						URI: doc.URI,
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      field.Name.Pos().Line - 1,
								Character: field.Name.Pos().Column - 1,
							},
							End: lsp.Position{
								Line:      field.Name.Pos().Line - 1,
								Character: field.Name.Pos().Column - 1 + len(word),
							},
						},
					}
				}
			}
		case *ast.InterfaceDecl:
			if d.Name.Value == word {
				return &lsp.Location{
					URI: doc.URI,
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      d.Name.Pos().Line - 1,
							Character: d.Name.Pos().Column - 1,
						},
						End: lsp.Position{
							Line:      d.Name.Pos().Line - 1,
							Character: d.Name.Pos().Column - 1 + len(word),
						},
					},
				}
			}
			// Check for method definitions
			for _, method := range d.Methods {
				if method.Name.Value == word {
					return &lsp.Location{
						URI: doc.URI,
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      method.Name.Pos().Line - 1,
								Character: method.Name.Pos().Column - 1,
							},
							End: lsp.Position{
								Line:      method.Name.Pos().Line - 1,
								Character: method.Name.Pos().Column - 1 + len(word),
							},
						},
					}
				}
			}
		}
	}

	return nil
}
