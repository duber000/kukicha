package main

import (
	"strings"
	"testing"

	"github.com/duber000/kukicha/internal/ast"
	"gopkg.in/yaml.v3"
)

func TestGenerateSkillMD_ProducesParseableYAML(t *testing.T) {
	skill := &ast.SkillDecl{
		Name:        &ast.Identifier{Value: "HelloSkill"},
		Description: "desc: has colon and \"quotes\"",
		Version:     "1.2.3",
	}

	functions := []FunctionSchema{
		{
			Name: "DoThing",
			Parameters: []ParameterSchema{
				{Name: "message", Type: "string", Default: "line1\nline2: value", HasDefault: true},
				{Name: "count", Type: "integer", Default: int64(3), HasDefault: true},
			},
		},
	}

	out := generateSkillMD(skill, functions)
	if !strings.HasPrefix(out, "---\n") || !strings.HasSuffix(out, "---\n") {
		t.Fatalf("expected frontmatter delimiters, got: %q", out)
	}

	content := strings.TrimPrefix(out, "---\n")
	content = strings.TrimSuffix(content, "---\n")

	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(content), &parsed); err != nil {
		t.Fatalf("generated YAML should be parseable: %v\n%s", err, out)
	}

	if parsed["name"] != "hello_skill" {
		t.Fatalf("unexpected name: %#v", parsed["name"])
	}
}

func TestDefaultValueToYAML(t *testing.T) {
	cases := []struct {
		name string
		expr ast.Expression
		want any
		ok   bool
	}{
		{name: "int", expr: &ast.IntegerLiteral{Value: 7}, want: int64(7), ok: true},
		{name: "float", expr: &ast.FloatLiteral{Value: 3.14}, want: 3.14, ok: true},
		{name: "string", expr: &ast.StringLiteral{Value: "x:y"}, want: "x:y", ok: true},
		{name: "bool", expr: &ast.BooleanLiteral{Value: true}, want: true, ok: true},
		{name: "unsupported", expr: &ast.Identifier{Value: "x"}, want: nil, ok: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := defaultValueToYAML(tc.expr)
			if ok != tc.ok {
				t.Fatalf("ok mismatch: got %v want %v", ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("value mismatch: got %#v want %#v", got, tc.want)
			}
		})
	}
}
