package lsp

import (
	"testing"

	"github.com/duber000/kukicha/internal/lexer"
	"github.com/sourcegraph/go-lsp"
)

func TestGetCompletions_ContainsKeywords(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, "func main()\n    x := 1\n", 1)

	doc := store.Get(uri)
	if doc == nil {
		t.Fatal("expected document")
	}

	items := s.getCompletions(doc, lsp.Position{Line: 1, Character: 0})

	// Should contain all keywords from the lexer
	keywords := lexer.Keywords()
	itemMap := make(map[string]bool)
	for _, item := range items {
		itemMap[item.Label] = true
	}

	for _, kw := range keywords {
		if !itemMap[kw] {
			t.Errorf("missing keyword completion: %s", kw)
		}
	}
}

func TestGetCompletions_ContainsBuiltins(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, "func main()\n    x := 1\n", 1)

	doc := store.Get(uri)
	items := s.getCompletions(doc, lsp.Position{Line: 1, Character: 0})

	itemMap := make(map[string]lsp.CompletionItem)
	for _, item := range items {
		itemMap[item.Label] = item
	}

	// Check that core builtins are present with correct kind
	builtinNames := []string{"print", "len", "append", "make", "close", "panic", "recover"}
	for _, name := range builtinNames {
		item, ok := itemMap[name]
		if !ok {
			t.Errorf("missing builtin completion: %s", name)
			continue
		}
		if item.Kind != lsp.CIKFunction {
			t.Errorf("expected %s to have kind Function, got %d", name, item.Kind)
		}
	}
}

func TestGetCompletions_ContainsPrimitiveTypes(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, "func main()\n    x := 1\n", 1)

	doc := store.Get(uri)
	items := s.getCompletions(doc, lsp.Position{Line: 1, Character: 0})

	itemMap := make(map[string]lsp.CompletionItem)
	for _, item := range items {
		itemMap[item.Label] = item
	}

	types := []string{"int", "string", "bool", "float64", "byte", "rune", "error", "any"}
	for _, ty := range types {
		item, ok := itemMap[ty]
		if !ok {
			t.Errorf("missing type completion: %s", ty)
			continue
		}
		if item.Kind != lsp.CIKTypeParameter {
			t.Errorf("expected %s to have kind TypeParameter, got %d", ty, item.Kind)
		}
	}
}

func TestGetCompletions_IncludesDocumentDeclarations(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `type Todo
    id int
    title string

func CreateTodo(title string) Todo
    return Todo{}
`, 1)

	doc := store.Get(uri)
	items := s.getCompletions(doc, lsp.Position{Line: 5, Character: 0})

	itemMap := make(map[string]lsp.CompletionItem)
	for _, item := range items {
		itemMap[item.Label] = item
	}

	// Should include the function declaration
	if item, ok := itemMap["CreateTodo"]; !ok {
		t.Error("missing function completion: CreateTodo")
	} else if item.Kind != lsp.CIKFunction {
		t.Errorf("expected CreateTodo to be Function kind, got %d", item.Kind)
	}

	// Should include the type declaration
	if item, ok := itemMap["Todo"]; !ok {
		t.Error("missing type completion: Todo")
	} else if item.Kind != lsp.CIKStruct {
		t.Errorf("expected Todo to be Struct kind, got %d", item.Kind)
	}
}

func TestGetCompletions_NilDocument(t *testing.T) {
	s := NewServer(nil, nil)

	// Asking for a document that doesn't exist should return empty
	doc := s.documents.Get("file:///nonexistent.kuki")
	if doc != nil {
		t.Fatal("expected nil document")
	}
}

func TestGetCompletions_IncludesInterfaceDecl(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `interface Reader
    Read(p list of byte) (int, error)
`, 1)

	doc := store.Get(uri)
	items := s.getCompletions(doc, lsp.Position{Line: 1, Character: 0})

	found := false
	for _, item := range items {
		if item.Label == "Reader" && item.Kind == lsp.CIKInterface {
			found = true
			break
		}
	}
	if !found {
		t.Error("missing interface completion: Reader")
	}
}

func TestGetDocumentSymbols_FunctionAndType(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `petiole main

type Todo
    id int
    title string

func CreateTodo(title string) Todo
    return Todo{}
`, 1)

	doc := store.Get(uri)
	symbols := s.getDocumentSymbols(doc)

	symbolNames := make(map[string]lsp.SymbolKind)
	for _, sym := range symbols {
		symbolNames[sym.Name] = sym.Kind
	}

	// Package
	if kind, ok := symbolNames["main"]; !ok {
		t.Error("missing package symbol: main")
	} else if kind != lsp.SKPackage {
		t.Errorf("expected main to be Package, got %d", kind)
	}

	// Function
	if kind, ok := symbolNames["CreateTodo"]; !ok {
		t.Error("missing function symbol: CreateTodo")
	} else if kind != lsp.SKFunction {
		t.Errorf("expected CreateTodo to be Function, got %d", kind)
	}

	// Type
	if kind, ok := symbolNames["Todo"]; !ok {
		t.Error("missing type symbol: Todo")
	} else if kind != lsp.SKStruct {
		t.Errorf("expected Todo to be Struct, got %d", kind)
	}

	// Fields
	if kind, ok := symbolNames["id"]; !ok {
		t.Error("missing field symbol: id")
	} else if kind != lsp.SKField {
		t.Errorf("expected id to be Field, got %d", kind)
	}
}

func TestGetDocumentSymbols_NilProgram(t *testing.T) {
	s := NewServer(nil, nil)
	doc := &Document{Program: nil}
	symbols := s.getDocumentSymbols(doc)

	if len(symbols) != 0 {
		t.Errorf("expected 0 symbols for nil program, got %d", len(symbols))
	}
}

func TestGetDocumentSymbols_InterfaceWithMethods(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `interface Storage
    Get(key string) (string, error)
    Set(key string, value string) error
`, 1)

	doc := store.Get(uri)
	symbols := s.getDocumentSymbols(doc)

	symbolNames := make(map[string]lsp.SymbolKind)
	for _, sym := range symbols {
		symbolNames[sym.Name] = sym.Kind
	}

	if kind, ok := symbolNames["Storage"]; !ok {
		t.Error("missing interface symbol: Storage")
	} else if kind != lsp.SKInterface {
		t.Errorf("expected Storage to be Interface, got %d", kind)
	}

	if kind, ok := symbolNames["Get"]; !ok {
		t.Error("missing method symbol: Get")
	} else if kind != lsp.SKMethod {
		t.Errorf("expected Get to be Method, got %d", kind)
	}

	if kind, ok := symbolNames["Set"]; !ok {
		t.Error("missing method symbol: Set")
	} else if kind != lsp.SKMethod {
		t.Errorf("expected Set to be Method, got %d", kind)
	}
}
