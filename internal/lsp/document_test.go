package lsp

import (
	"testing"

	"github.com/sourcegraph/go-lsp"
)

func TestDocumentStoreGetReturnsSnapshot(t *testing.T) {
	store := NewDocumentStore()
	uri := lsp.DocumentURI("file:///tmp/test.kuki")

	opened := store.Open(uri, "func A()\n    x := 1\n", 1)
	got := store.Get(uri)
	if got == nil {
		t.Fatal("expected document from store")
	}
	if got == opened {
		t.Fatal("expected Get to return a cloned snapshot, got same pointer")
	}

	got.Content = "mutated"
	got.Lines[0] = "mutated-line"

	again := store.Get(uri)
	if again.Content != "func A()\n    x := 1\n" {
		t.Fatalf("expected store content to remain unchanged, got %q", again.Content)
	}
	if again.Lines[0] != "func A()" {
		t.Fatalf("expected store lines to remain unchanged, got %q", again.Lines[0])
	}
}

