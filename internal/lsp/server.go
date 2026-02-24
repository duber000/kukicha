package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

// Server implements the Kukicha Language Server Protocol
type Server struct {
	conn      *jsonrpc2.Conn
	reader    io.Reader
	writer    io.Writer
	documents *DocumentStore
	mu        sync.RWMutex
}

// NewServer creates a new LSP server
func NewServer(reader io.Reader, writer io.Writer) *Server {
	return &Server{
		reader:    reader,
		writer:    writer,
		documents: NewDocumentStore(),
	}
}

// Run starts the LSP server and processes requests
func (s *Server) Run(ctx context.Context) error {
	stream := jsonrpc2.NewBufferedStream(
		&readWriteCloser{s.reader, s.writer},
		jsonrpc2.VSCodeObjectCodec{},
	)

	s.conn = jsonrpc2.NewConn(ctx, stream, s)

	<-s.conn.DisconnectNotify()
	return nil
}

// Handle implements jsonrpc2.Handler
func (s *Server) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	log.Printf("Received request: %s", req.Method)

	result, err := s.handleRequest(ctx, req)
	if err != nil {
		if !req.Notif {
			if respErr := conn.ReplyWithError(ctx, req.ID, toJSONRPCError(err)); respErr != nil {
				log.Printf("Error sending error response: %v", respErr)
			}
		}
		return
	}

	if !req.Notif {
		if err := conn.Reply(ctx, req.ID, result); err != nil {
			log.Printf("Error sending response: %v", err)
		}
	}
}

func (s *Server) handleRequest(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, req)
	case "initialized":
		return nil, nil
	case "shutdown":
		return nil, nil
	case "exit":
		return nil, nil
	case "textDocument/didOpen":
		return s.handleDidOpen(ctx, req)
	case "textDocument/didChange":
		return s.handleDidChange(ctx, req)
	case "textDocument/didSave":
		return s.handleDidSave(ctx, req)
	case "textDocument/didClose":
		return s.handleDidClose(ctx, req)
	case "textDocument/hover":
		return s.handleHover(ctx, req)
	case "textDocument/definition":
		return s.handleDefinition(ctx, req)
	case "textDocument/completion":
		return s.handleCompletion(ctx, req)
	case "textDocument/documentSymbol":
		return s.handleDocumentSymbol(ctx, req)
	default:
		return nil, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeMethodNotFound,
			Message: fmt.Sprintf("method not found: %s", req.Method),
		}
	}
}

func toJSONRPCError(err error) *jsonrpc2.Error {
	var rpcErr *jsonrpc2.Error
	if errors.As(err, &rpcErr) {
		return rpcErr
	}
	return &jsonrpc2.Error{
		Code:    jsonrpc2.CodeInternalError,
		Message: err.Error(),
	}
}

func (s *Server) handleInitialize(ctx context.Context, req *jsonrpc2.Request) (*lsp.InitializeResult, error) {
	log.Println("Handling initialize request")

	result := &lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: &lsp.TextDocumentSyncOptionsOrKind{
				Options: &lsp.TextDocumentSyncOptions{
					OpenClose: true,
					Change:    lsp.TDSKFull,
					Save: &lsp.SaveOptions{
						IncludeText: true,
					},
				},
			},
			HoverProvider:      true,
			DefinitionProvider: true,
			CompletionProvider: &lsp.CompletionOptions{
				TriggerCharacters: []string{".", ":"},
			},
			DocumentSymbolProvider: true,
		},
	}

	return result, nil
}

func (s *Server) handleDidOpen(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	log.Printf("Document opened: %s", params.TextDocument.URI)

	s.documents.Open(params.TextDocument.URI, params.TextDocument.Text, params.TextDocument.Version)

	// Analyze and publish diagnostics
	s.publishDiagnostics(ctx, params.TextDocument.URI)

	return nil, nil
}

func (s *Server) handleDidChange(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	// Apply changes (we're using full sync, so just take the whole content)
	if len(params.ContentChanges) > 0 {
		s.documents.Update(params.TextDocument.URI, params.ContentChanges[0].Text, params.TextDocument.Version)
	}

	// Analyze and publish diagnostics
	s.publishDiagnostics(ctx, params.TextDocument.URI)

	return nil, nil
}

func (s *Server) handleDidSave(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	var params lsp.DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	log.Printf("Document saved: %s", params.TextDocument.URI)

	// Re-analyze on save
	s.publishDiagnostics(ctx, params.TextDocument.URI)

	return nil, nil
}

func (s *Server) handleDidClose(ctx context.Context, req *jsonrpc2.Request) (any, error) {
	var params lsp.DidCloseTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	log.Printf("Document closed: %s", params.TextDocument.URI)

	s.documents.Close(params.TextDocument.URI)

	// Clear diagnostics
	s.conn.Notify(ctx, "textDocument/publishDiagnostics", &lsp.PublishDiagnosticsParams{
		URI:         params.TextDocument.URI,
		Diagnostics: []lsp.Diagnostic{},
	})

	return nil, nil
}

// readWriteCloser wraps io.Reader and io.Writer to implement io.ReadWriteCloser
type readWriteCloser struct {
	io.Reader
	io.Writer
}

func (rwc *readWriteCloser) Close() error {
	return nil
}
