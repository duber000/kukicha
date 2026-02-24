package lsp

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/jsonrpc2"
)

func TestHandleRequestUnknownMethodReturnsMethodNotFound(t *testing.T) {
	s := NewServer(nil, nil)
	req := &jsonrpc2.Request{Method: "unknown/method"}

	_, err := s.handleRequest(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for unknown method")
	}

	var rpcErr *jsonrpc2.Error
	if !errors.As(err, &rpcErr) {
		t.Fatalf("expected *jsonrpc2.Error, got %T", err)
	}
	if rpcErr.Code != jsonrpc2.CodeMethodNotFound {
		t.Fatalf("expected code %d, got %d", jsonrpc2.CodeMethodNotFound, rpcErr.Code)
	}
}

