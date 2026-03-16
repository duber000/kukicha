package codegen

import (
	"testing"

	"github.com/duber000/kukicha/internal/semantic"
)

func TestTypeInfoToGoString(t *testing.T) {
	prog := mustParseProgram(t, "func main()\n    x := 1\n")
	gen := New(prog)

	tests := []struct {
		name     string
		input    *semantic.TypeInfo
		expected string
	}{
		{"nil", nil, "any"},
		{"int", &semantic.TypeInfo{Kind: semantic.TypeKindInt}, "int"},
		{"float64", &semantic.TypeInfo{Kind: semantic.TypeKindFloat}, "float64"},
		{"string", &semantic.TypeInfo{Kind: semantic.TypeKindString}, "string"},
		{"bool", &semantic.TypeInfo{Kind: semantic.TypeKindBool}, "bool"},
		{"list of int", &semantic.TypeInfo{
			Kind:        semantic.TypeKindList,
			ElementType: &semantic.TypeInfo{Kind: semantic.TypeKindInt},
		}, "[]int"},
		{"list of string", &semantic.TypeInfo{
			Kind:        semantic.TypeKindList,
			ElementType: &semantic.TypeInfo{Kind: semantic.TypeKindString},
		}, "[]string"},
		{"map of string to int", &semantic.TypeInfo{
			Kind:      semantic.TypeKindMap,
			KeyType:   &semantic.TypeInfo{Kind: semantic.TypeKindString},
			ValueType: &semantic.TypeInfo{Kind: semantic.TypeKindInt},
		}, "map[string]int"},
		{"channel of string", &semantic.TypeInfo{
			Kind:        semantic.TypeKindChannel,
			ElementType: &semantic.TypeInfo{Kind: semantic.TypeKindString},
		}, "chan string"},
		{"reference int", &semantic.TypeInfo{
			Kind:        semantic.TypeKindReference,
			ElementType: &semantic.TypeInfo{Kind: semantic.TypeKindInt},
		}, "*int"},
		{"named type", &semantic.TypeInfo{
			Kind: semantic.TypeKindNamed,
			Name: "User",
		}, "User"},
		{"named type with package", &semantic.TypeInfo{
			Kind: semantic.TypeKindNamed,
			Name: "http.Request",
		}, "http.Request"},
		{"function type no return", &semantic.TypeInfo{
			Kind:   semantic.TypeKindFunction,
			Params: []*semantic.TypeInfo{{Kind: semantic.TypeKindString}},
		}, "func(string)"},
		{"function type single return", &semantic.TypeInfo{
			Kind:    semantic.TypeKindFunction,
			Params:  []*semantic.TypeInfo{{Kind: semantic.TypeKindString}},
			Returns: []*semantic.TypeInfo{{Kind: semantic.TypeKindBool}},
		}, "func(string) bool"},
		{"function type multi return", &semantic.TypeInfo{
			Kind:   semantic.TypeKindFunction,
			Params: []*semantic.TypeInfo{{Kind: semantic.TypeKindString}},
			Returns: []*semantic.TypeInfo{
				{Kind: semantic.TypeKindInt},
				{Kind: semantic.TypeKindNamed, Name: "error"},
			},
		}, "func(string) (int, error)"},
		{"nested list of map", &semantic.TypeInfo{
			Kind: semantic.TypeKindList,
			ElementType: &semantic.TypeInfo{
				Kind:      semantic.TypeKindMap,
				KeyType:   &semantic.TypeInfo{Kind: semantic.TypeKindString},
				ValueType: &semantic.TypeInfo{Kind: semantic.TypeKindInt},
			},
		}, "[]map[string]int"},
		{"unknown kind", &semantic.TypeInfo{Kind: semantic.TypeKindUnknown}, "any"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gen.typeInfoToGoString(tt.input)
			if got != tt.expected {
				t.Errorf("typeInfoToGoString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTypeInfoToGoStringWithPkgAlias(t *testing.T) {
	prog := mustParseProgram(t, "func main()\n    x := 1\n")
	gen := New(prog)
	gen.pkgAliases["json"] = "kukijson"

	ti := &semantic.TypeInfo{
		Kind: semantic.TypeKindNamed,
		Name: "json.Decoder",
	}

	got := gen.typeInfoToGoString(ti)
	if got != "kukijson.Decoder" {
		t.Errorf("expected kukijson.Decoder with alias, got %q", got)
	}
}
