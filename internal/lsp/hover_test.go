package lsp

import (
	"strings"
	"testing"

	"github.com/sourcegraph/go-lsp"
)

func TestGetHoverContent_BuiltinFunction(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, "func main()\n    print(42)\n", 1)

	doc := store.Get(uri)
	content := s.getHoverContent(doc, "print", lsp.Position{Line: 1, Character: 4})

	if content == "" {
		t.Fatal("expected hover content for builtin 'print'")
	}
	if !strings.Contains(content, "print") {
		t.Errorf("hover content should mention 'print', got: %s", content)
	}
}

func TestGetHoverContent_FunctionDecl(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `func Add(a int, b int) int
    return a + b
`, 1)

	doc := store.Get(uri)
	content := s.getHoverContent(doc, "Add", lsp.Position{Line: 0, Character: 5})

	if content == "" {
		t.Fatal("expected hover content for 'Add'")
	}
	if !strings.Contains(content, "Add") {
		t.Errorf("expected 'Add' in hover, got: %s", content)
	}
	if !strings.Contains(content, "int") {
		t.Errorf("expected return type 'int' in hover, got: %s", content)
	}
}

func TestGetHoverContent_TypeDecl(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `type Todo
    id int
    title string
`, 1)

	doc := store.Get(uri)
	content := s.getHoverContent(doc, "Todo", lsp.Position{Line: 0, Character: 5})

	if content == "" {
		t.Fatal("expected hover content for 'Todo'")
	}
	if !strings.Contains(content, "Todo") {
		t.Errorf("expected 'Todo' in hover, got: %s", content)
	}
	if !strings.Contains(content, "id") {
		t.Errorf("expected field 'id' in hover, got: %s", content)
	}
}

func TestGetHoverContent_InterfaceDecl(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `interface Reader
    Read(p list of byte) (int, error)
`, 1)

	doc := store.Get(uri)
	content := s.getHoverContent(doc, "Reader", lsp.Position{Line: 0, Character: 10})

	if content == "" {
		t.Fatal("expected hover content for 'Reader'")
	}
	if !strings.Contains(content, "Reader") {
		t.Errorf("expected 'Reader' in hover, got: %s", content)
	}
}

func TestGetHoverContent_UnknownSymbol(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, "func main()\n    x := 1\n", 1)

	doc := store.Get(uri)
	content := s.getHoverContent(doc, "nonexistent", lsp.Position{Line: 0, Character: 0})

	if content != "" {
		t.Errorf("expected empty hover for unknown symbol, got: %s", content)
	}
}

func TestGetHoverContent_NilProgram(t *testing.T) {
	s := NewServer(nil, nil)
	doc := &Document{Program: nil}

	content := s.getHoverContent(doc, "anything", lsp.Position{Line: 0, Character: 0})
	if content != "" {
		t.Errorf("expected empty hover for nil program, got: %s", content)
	}
}

func TestGetHoverContent_EmptyWord(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, "func main()\n    x := 1\n", 1)

	doc := store.Get(uri)
	content := s.getHoverContent(doc, "", lsp.Position{Line: 0, Character: 0})

	// Empty word should match no builtin and no declaration
	if content != "" {
		t.Errorf("expected empty hover for empty word, got: %s", content)
	}
}

func TestFormatFunctionDecl_WithReceiver(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `type Counter
    value int

func Increment on c reference Counter
    c.value = c.value + 1
`, 1)

	doc := store.Get(uri)
	content := s.getHoverContent(doc, "Increment", lsp.Position{Line: 3, Character: 5})

	if content == "" {
		t.Fatal("expected hover content for method 'Increment'")
	}
	if !strings.Contains(content, "Increment") {
		t.Errorf("expected 'Increment' in hover, got: %s", content)
	}
}

func TestFormatTypeAnnotation_AllTypes(t *testing.T) {
	tests := []struct {
		name   string
		source string
		hover  string
		substr string
	}{
		{
			name:   "list type",
			source: "func Get() list of string\n    return list of string{}\n",
			hover:  "Get",
			substr: "list of string",
		},
		{
			name:   "map type",
			source: "func Get() map of string to int\n    return map of string to int{}\n",
			hover:  "Get",
			substr: "map of string to int",
		},
		{
			name:   "reference type",
			source: "type Wrapper\n    data reference string\n",
			hover:  "Wrapper",
			substr: "reference string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewServer(nil, nil)
			store := s.documents
			uri := lsp.DocumentURI("file:///tmp/test.kuki")
			store.Open(uri, tt.source, 1)

			doc := store.Get(uri)
			content := s.getHoverContent(doc, tt.hover, lsp.Position{Line: 0, Character: 5})

			if content == "" {
				t.Fatalf("expected hover content for %s", tt.hover)
			}
			if !strings.Contains(content, tt.substr) {
				t.Errorf("expected %q in hover, got: %s", tt.substr, content)
			}
		})
	}
}

func TestLookupBuiltin_AllBuiltins(t *testing.T) {
	for _, b := range builtins {
		result := lookupBuiltin(b.Name)
		if result == "" {
			t.Errorf("lookupBuiltin(%q) returned empty", b.Name)
		}
		if !strings.Contains(result, b.Signature) {
			t.Errorf("lookupBuiltin(%q) missing signature, got: %s", b.Name, result)
		}
		if !strings.Contains(result, b.Doc) {
			t.Errorf("lookupBuiltin(%q) missing doc, got: %s", b.Name, result)
		}
	}
}

func TestLookupBuiltin_Unknown(t *testing.T) {
	result := lookupBuiltin("nonexistent_builtin")
	if result != "" {
		t.Errorf("expected empty for unknown builtin, got: %s", result)
	}
}
