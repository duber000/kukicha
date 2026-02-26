package lsp

import (
	"testing"

	"github.com/sourcegraph/go-lsp"
)

func TestFindDefinition_FunctionDecl(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `func Add(a int, b int) int
    return a + b

func main()
    x := Add(1, 2)
`, 1)

	doc := store.Get(uri)
	loc := s.findDefinition(doc, "Add")

	if loc == nil {
		t.Fatal("expected definition location for 'Add'")
	}
	if loc.URI != uri {
		t.Errorf("expected URI %s, got %s", uri, loc.URI)
	}
	// "Add" is declared on line 1 (0-indexed: 0)
	if loc.Range.Start.Line != 0 {
		t.Errorf("expected definition on line 0, got %d", loc.Range.Start.Line)
	}
}

func TestFindDefinition_TypeDecl(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `type Todo
    id int
    title string
`, 1)

	doc := store.Get(uri)
	loc := s.findDefinition(doc, "Todo")

	if loc == nil {
		t.Fatal("expected definition location for 'Todo'")
	}
	if loc.Range.Start.Line != 0 {
		t.Errorf("expected definition on line 0, got %d", loc.Range.Start.Line)
	}
}

func TestFindDefinition_FieldDecl(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `type Todo
    id int
    title string
`, 1)

	doc := store.Get(uri)
	loc := s.findDefinition(doc, "title")

	if loc == nil {
		t.Fatal("expected definition location for field 'title'")
	}
	// "title" is on the 3rd line (0-indexed: 2)
	if loc.Range.Start.Line != 2 {
		t.Errorf("expected field definition on line 2, got %d", loc.Range.Start.Line)
	}
}

func TestFindDefinition_InterfaceDecl(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `interface Storage
    Get(key string) (string, error)
    Set(key string, value string) error
`, 1)

	doc := store.Get(uri)
	loc := s.findDefinition(doc, "Storage")

	if loc == nil {
		t.Fatal("expected definition location for 'Storage'")
	}
	if loc.Range.Start.Line != 0 {
		t.Errorf("expected definition on line 0, got %d", loc.Range.Start.Line)
	}
}

func TestFindDefinition_InterfaceMethod(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `interface Storage
    Get(key string) (string, error)
    Set(key string, value string) error
`, 1)

	doc := store.Get(uri)
	loc := s.findDefinition(doc, "Get")

	if loc == nil {
		t.Fatal("expected definition location for interface method 'Get'")
	}
	// "Get" is on line 2 (0-indexed: 1)
	if loc.Range.Start.Line != 1 {
		t.Errorf("expected method definition on line 1, got %d", loc.Range.Start.Line)
	}
}

func TestFindDefinition_UnknownSymbol(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, "func main()\n    x := 1\n", 1)

	doc := store.Get(uri)
	loc := s.findDefinition(doc, "nonexistent")

	if loc != nil {
		t.Errorf("expected nil for unknown symbol, got %+v", loc)
	}
}

func TestFindDefinition_NilProgram(t *testing.T) {
	s := NewServer(nil, nil)
	doc := &Document{Program: nil}

	loc := s.findDefinition(doc, "anything")
	if loc != nil {
		t.Errorf("expected nil for nil program, got %+v", loc)
	}
}

func TestFindDefinition_MultipleDeclarations(t *testing.T) {
	s := NewServer(nil, nil)
	store := s.documents
	uri := lsp.DocumentURI("file:///tmp/test.kuki")
	store.Open(uri, `type User
    name string

type Post
    title string
    author string

func GetUser() User
    return User{}
`, 1)

	doc := store.Get(uri)

	// Should find User at line 0
	userLoc := s.findDefinition(doc, "User")
	if userLoc == nil {
		t.Fatal("expected definition for 'User'")
	}
	if userLoc.Range.Start.Line != 0 {
		t.Errorf("expected User at line 0, got %d", userLoc.Range.Start.Line)
	}

	// Should find Post at line 3
	postLoc := s.findDefinition(doc, "Post")
	if postLoc == nil {
		t.Fatal("expected definition for 'Post'")
	}
	if postLoc.Range.Start.Line != 3 {
		t.Errorf("expected Post at line 3, got %d", postLoc.Range.Start.Line)
	}

	// Should find author field at line 5
	authorLoc := s.findDefinition(doc, "author")
	if authorLoc == nil {
		t.Fatal("expected definition for field 'author'")
	}
	if authorLoc.Range.Start.Line != 5 {
		t.Errorf("expected author at line 5, got %d", authorLoc.Range.Start.Line)
	}
}
