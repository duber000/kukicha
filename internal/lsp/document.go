package lsp

import (
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/duber000/kukicha/internal/ast"
	"github.com/duber000/kukicha/internal/parser"
	"github.com/duber000/kukicha/internal/semantic"
	"github.com/sourcegraph/go-lsp"
)

// Document represents an open document in the LSP
type Document struct {
	URI     lsp.DocumentURI
	Content string
	Version int

	// Cached analysis results
	Program     *ast.Program
	SymbolTable *semantic.SymbolTable
	Errors      []error
	Lines       []string
}

// DocumentStore manages all open documents
type DocumentStore struct {
	documents map[lsp.DocumentURI]*Document
	mu        sync.RWMutex
}

// NewDocumentStore creates a new document store
func NewDocumentStore() *DocumentStore {
	return &DocumentStore{
		documents: make(map[lsp.DocumentURI]*Document),
	}
}

// Open adds a new document to the store
func (ds *DocumentStore) Open(uri lsp.DocumentURI, content string, version int) *Document {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	doc := &Document{
		URI:     uri,
		Content: content,
		Version: version,
		Lines:   strings.Split(content, "\n"),
	}

	// Parse and analyze the document
	doc.analyze()

	ds.documents[uri] = doc
	return doc
}

// Update updates an existing document
func (ds *DocumentStore) Update(uri lsp.DocumentURI, content string, version int) *Document {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	doc, exists := ds.documents[uri]
	if !exists {
		doc = &Document{URI: uri}
		ds.documents[uri] = doc
	}

	doc.Content = content
	doc.Version = version
	doc.Lines = strings.Split(content, "\n")

	// Re-analyze the document
	doc.analyze()

	return doc
}

// Close removes a document from the store
func (ds *DocumentStore) Close(uri lsp.DocumentURI) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	delete(ds.documents, uri)
}

// Get retrieves a document by URI
func (ds *DocumentStore) Get(uri lsp.DocumentURI) *Document {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.documents[uri]
}

// analyze parses and performs semantic analysis on the document
func (doc *Document) analyze() {
	// Extract filename from URI
	filename := uriToFilename(doc.URI)

	// Parse the content (parser handles lexing internally)
	p, err := parser.New(doc.Content, filename)
	if err != nil {
		doc.Errors = []error{err}
		doc.Program = nil
		return
	}

	program, parseErrors := p.Parse()
	doc.Program = program
	doc.Errors = parseErrors

	// Only run semantic analysis if parsing succeeded
	if len(doc.Errors) == 0 && program != nil {
		analyzer := semantic.New(program)
		semanticErrors := analyzer.Analyze()
		doc.Errors = append(doc.Errors, semanticErrors...)
	}
}

// GetLineContent returns the content of a specific line (0-indexed)
func (doc *Document) GetLineContent(line int) string {
	if line < 0 || line >= len(doc.Lines) {
		return ""
	}
	return doc.Lines[line]
}

// PositionToOffset converts an LSP position to a byte offset
func (doc *Document) PositionToOffset(pos lsp.Position) int {
	offset := 0
	for i := 0; i < int(pos.Line) && i < len(doc.Lines); i++ {
		offset += len(doc.Lines[i]) + 1 // +1 for newline
	}
	if int(pos.Line) < len(doc.Lines) {
		offset += utf16PosToByteOffset(doc.Lines[int(pos.Line)], int(pos.Character))
	}
	return offset
}

// OffsetToPosition converts a byte offset to an LSP position
func (doc *Document) OffsetToPosition(offset int) lsp.Position {
	currentOffset := 0
	for line, content := range doc.Lines {
		lineEnd := currentOffset + len(content) + 1 // +1 for newline
		if offset < lineEnd {
			byteInLine := offset - currentOffset
			if byteInLine < 0 {
				byteInLine = 0
			}
			if byteInLine > len(content) {
				byteInLine = len(content)
			}
			return lsp.Position{
				Line:      line,
				Character: byteOffsetToUTF16Pos(content, byteInLine),
			}
		}
		currentOffset = lineEnd
	}
	// Return end of document
	return lsp.Position{
		Line:      len(doc.Lines) - 1,
		Character: byteOffsetToUTF16Pos(doc.Lines[len(doc.Lines)-1], len(doc.Lines[len(doc.Lines)-1])),
	}
}

// GetWordAtPosition returns the word at the given position
func (doc *Document) GetWordAtPosition(pos lsp.Position) string {
	line := doc.GetLineContent(int(pos.Line))
	if line == "" {
		return ""
	}

	col := utf16PosToByteOffset(line, int(pos.Character))
	if col >= len(line) {
		col = len(line) - 1
	}
	if col < 0 {
		return ""
	}

	// Find word boundaries
	start := col
	for start > 0 && isIdentifierChar(line[start-1]) {
		start--
	}

	end := col
	for end < len(line) && isIdentifierChar(line[end]) {
		end++
	}

	return line[start:end]
}

// isIdentifierChar returns true if the character can be part of an identifier
func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_'
}

func uriToFilename(uri lsp.DocumentURI) string {
	raw := string(uri)
	parsed, err := url.Parse(raw)
	if err == nil && parsed.Scheme == "file" {
		path := parsed.Path
		if parsed.Host != "" {
			path = "//" + parsed.Host + path
		}
		return filepath.FromSlash(path)
	}
	return strings.TrimPrefix(raw, "file://")
}

func utf16PosToByteOffset(line string, utf16Pos int) int {
	if utf16Pos <= 0 {
		return 0
	}
	utf16Count := 0
	for idx, r := range line {
		utf16Count += utf16RuneLen(r)
		if utf16Count > utf16Pos {
			return idx
		}
	}
	return len(line)
}

func byteOffsetToUTF16Pos(line string, byteOffset int) int {
	if byteOffset <= 0 {
		return 0
	}
	if byteOffset > len(line) {
		byteOffset = len(line)
	}
	utf16Count := 0
	for idx, r := range line {
		if idx >= byteOffset {
			break
		}
		utf16Count += utf16RuneLen(r)
	}
	return utf16Count
}

func utf16RuneLen(r rune) int {
	if r <= 0xFFFF {
		return 1
	}
	return 2
}
