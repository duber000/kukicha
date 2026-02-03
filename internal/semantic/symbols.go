package semantic

import (
	"fmt"

	"github.com/duber000/kukicha/internal/ast"
)

// SymbolKind represents the kind of symbol
type SymbolKind int

const (
	SymbolVariable SymbolKind = iota
	SymbolParameter
	SymbolFunction
	SymbolType
	SymbolInterface
)

func (sk SymbolKind) String() string {
	switch sk {
	case SymbolVariable:
		return "variable"
	case SymbolParameter:
		return "parameter"
	case SymbolFunction:
		return "function"
	case SymbolType:
		return "type"
	case SymbolInterface:
		return "interface"
	default:
		return "unknown"
	}
}

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name     string
	Kind     SymbolKind
	Type     *TypeInfo
	Defined  ast.Position
	Mutable  bool
	Exported bool
}

// TypeKind represents the kind of type
type TypeKind int

const (
	TypeKindUnknown TypeKind = iota
	TypeKindInt
	TypeKindFloat
	TypeKindString
	TypeKindBool
	TypeKindList
	TypeKindMap
	TypeKindChannel
	TypeKindReference
	TypeKindFunction
	TypeKindStruct
	TypeKindInterface
	TypeKindNamed
	TypeKindPlaceholder // For generic type placeholders (element, item, etc.)
)

func (tk TypeKind) String() string {
	switch tk {
	case TypeKindUnknown:
		return "unknown"
	case TypeKindInt:
		return "int"
	case TypeKindFloat:
		return "float"
	case TypeKindString:
		return "string"
	case TypeKindBool:
		return "bool"
	case TypeKindList:
		return "list"
	case TypeKindMap:
		return "map"
	case TypeKindChannel:
		return "channel"
	case TypeKindReference:
		return "reference"
	case TypeKindFunction:
		return "function"
	case TypeKindStruct:
		return "struct"
	case TypeKindInterface:
		return "interface"
	case TypeKindNamed:
		return "named"
	case TypeKindPlaceholder:
		return "placeholder"
	default:
		return "unknown"
	}
}

// TypeInfo represents type information
type TypeInfo struct {
	Kind         TypeKind
	Name         string      // For named types and placeholders
	ElementType  *TypeInfo   // For lists, channels, references
	KeyType      *TypeInfo   // For maps
	ValueType    *TypeInfo   // For maps
	Params       []*TypeInfo // For functions
	Returns      []*TypeInfo // For functions
	Constraint   string      // For placeholders: "any", "comparable", "cmp.Ordered"
	Variadic     bool        // For functions: true if last param is variadic
	ParamNames   []string    // For functions: parameter names (for named argument validation)
	DefaultCount int         // For functions: number of parameters with default values
}

func (ti *TypeInfo) String() string {
	if ti == nil {
		return "nil"
	}

	switch ti.Kind {
	case TypeKindNamed:
		return ti.Name
	case TypeKindPlaceholder:
		return ti.Name // Return the placeholder name (element, item, etc.)
	case TypeKindList:
		if ti.ElementType != nil {
			return fmt.Sprintf("list of %s", ti.ElementType)
		}
		return "list"
	case TypeKindMap:
		if ti.KeyType != nil && ti.ValueType != nil {
			return fmt.Sprintf("map of %s to %s", ti.KeyType, ti.ValueType)
		}
		return "map"
	case TypeKindChannel:
		if ti.ElementType != nil {
			return fmt.Sprintf("channel of %s", ti.ElementType)
		}
		return "channel"
	case TypeKindReference:
		if ti.ElementType != nil {
			return fmt.Sprintf("reference %s", ti.ElementType)
		}
		return "reference"
	case TypeKindFunction:
		params := ""
		for i, p := range ti.Params {
			if i > 0 {
				params += ", "
			}
			params += p.String()
		}
		returns := ""
		for i, r := range ti.Returns {
			if i > 0 {
				returns += ", "
			}
			returns += r.String()
		}
		if returns != "" {
			return fmt.Sprintf("func(%s) %s", params, returns)
		}
		return fmt.Sprintf("func(%s)", params)
	default:
		return ti.Kind.String()
	}
}

// Scope represents a lexical scope
type Scope struct {
	parent  *Scope
	symbols map[string]*Symbol
}

// NewScope creates a new scope
func NewScope(parent *Scope) *Scope {
	return &Scope{
		parent:  parent,
		symbols: make(map[string]*Symbol),
	}
}

// Define adds a symbol to the current scope
func (s *Scope) Define(symbol *Symbol) error {
	if _, exists := s.symbols[symbol.Name]; exists {
		return fmt.Errorf("identifier '%s' already declared in this scope", symbol.Name)
	}
	s.symbols[symbol.Name] = symbol
	return nil
}

// Resolve looks up a symbol in the current scope and parent scopes
func (s *Scope) Resolve(name string) *Symbol {
	if symbol, ok := s.symbols[name]; ok {
		return symbol
	}
	if s.parent != nil {
		return s.parent.Resolve(name)
	}
	return nil
}

// SymbolTable manages scopes and symbols
type SymbolTable struct {
	scopes []*Scope
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		scopes: []*Scope{NewScope(nil)}, // Global scope
	}
}

// CurrentScope returns the current scope
func (st *SymbolTable) CurrentScope() *Scope {
	if len(st.scopes) == 0 {
		return nil
	}
	return st.scopes[len(st.scopes)-1]
}

// EnterScope creates a new scope
func (st *SymbolTable) EnterScope() {
	newScope := NewScope(st.CurrentScope())
	st.scopes = append(st.scopes, newScope)
}

// ExitScope removes the current scope
func (st *SymbolTable) ExitScope() {
	if len(st.scopes) > 1 {
		st.scopes = st.scopes[:len(st.scopes)-1]
	}
}

// Define adds a symbol to the current scope
func (st *SymbolTable) Define(symbol *Symbol) error {
	return st.CurrentScope().Define(symbol)
}

// Resolve looks up a symbol
func (st *SymbolTable) Resolve(name string) *Symbol {
	return st.CurrentScope().Resolve(name)
}
