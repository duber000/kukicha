package codegen

import (
	"strings"
	"testing"
)

func TestGenerateInterfaceDecl(t *testing.T) {
	input := `interface Reader
    Read(buf list of byte) (int, error)
`

	output := generateSource(t, input)

	if !strings.Contains(output, "type Reader interface {") {
		t.Errorf("expected interface declaration, got:\n%s", output)
	}
	if !strings.Contains(output, "Read(buf []byte) (int, error)") {
		t.Errorf("expected Read method signature, got:\n%s", output)
	}
}

func TestGenerateInterfaceMultipleMethods(t *testing.T) {
	input := `interface Storage
    Get(key string) (string, error)
    Set(key string, value string) error
    Delete(key string) error
`

	output := generateSource(t, input)

	if !strings.Contains(output, "type Storage interface {") {
		t.Errorf("expected Storage interface, got:\n%s", output)
	}
	if !strings.Contains(output, "Get(key string) (string, error)") {
		t.Errorf("expected Get method, got:\n%s", output)
	}
	if !strings.Contains(output, "Set(key string, value string) error") {
		t.Errorf("expected Set method, got:\n%s", output)
	}
	if !strings.Contains(output, "Delete(key string) error") {
		t.Errorf("expected Delete method, got:\n%s", output)
	}
}

func TestGenerateGlobalVarDecl(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "typed with value",
			input:    "var AppName string = \"myapp\"\n\nfunc main()\n    x := AppName\n",
			expected: `var AppName string = "myapp"`,
		},
		{
			name:     "typed without value",
			input:    "var counter int\n\nfunc main()\n    x := counter\n",
			expected: "var counter int",
		},
		{
			name:     "inferred type",
			input:    "var limit = 100\n\nfunc main()\n    x := limit\n",
			expected: "var limit = 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := generateSource(t, tt.input)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected %q, got:\n%s", tt.expected, output)
			}
		})
	}
}

func TestGenerateMethodDecl(t *testing.T) {
	input := `type User
    name string

func GetName on u User string
    return u.name
`

	output := generateSource(t, input)

	if !strings.Contains(output, "func (u User) GetName() string {") {
		t.Errorf("expected method with receiver, got:\n%s", output)
	}
}

func TestGeneratePointerReceiverMethod(t *testing.T) {
	input := `type Counter
    value int

func Increment on c reference Counter
    c.value = c.value + 1
`

	output := generateSource(t, input)

	if !strings.Contains(output, "func (c *Counter) Increment() {") {
		t.Errorf("expected pointer receiver method, got:\n%s", output)
	}
}

func TestGenerateVariadicFunction(t *testing.T) {
	input := `func Sum(many numbers int) int
    total := 0
    for n in numbers
        total = total + n
    return total
`

	output := generateSource(t, input)

	if !strings.Contains(output, "func Sum(numbers ...int) int {") {
		t.Errorf("expected variadic parameter, got:\n%s", output)
	}
}

func TestGenerateTypeAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "list of string",
			input:    "func f(items list of string)\n    x := items\n",
			expected: "func f(items []string)",
		},
		{
			name:     "map of string to int",
			input:    "func f(m map of string to int)\n    x := m\n",
			expected: "func f(m map[string]int)",
		},
		{
			name:     "reference type",
			input:    "func f(p reference int)\n    x := p\n",
			expected: "func f(p *int)",
		},
		{
			name:     "channel type",
			input:    "func f(ch channel of string)\n    x := ch\n",
			expected: "func f(ch chan string)",
		},
		{
			name:     "nested list of list",
			input:    "func f(m list of list of int)\n    x := m\n",
			expected: "func f(m [][]int)",
		},
		{
			name:     "function type param",
			input:    "func apply(f func(int) string)\n    x := f\n",
			expected: "func apply(f func(int) string)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := generateSource(t, tt.input)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected %q, got:\n%s", tt.expected, output)
			}
		})
	}
}

func TestGenerateReturnTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single return",
			input:    "func f() string\n    return \"hello\"\n",
			expected: "func f() string {",
		},
		{
			name:     "multiple returns",
			input:    "func f() (int, error)\n    return 0, empty\n",
			expected: "func f() (int, error) {",
		},
		{
			name:     "no return",
			input:    "func f()\n    x := 1\n",
			expected: "func f() {",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := generateSource(t, tt.input)
			if !strings.Contains(output, tt.expected) {
				t.Errorf("expected %q, got:\n%s", tt.expected, output)
			}
		})
	}
}

func TestGenerateTypeAlias(t *testing.T) {
	input := "type Handler func(string) error\n"

	output := generateSource(t, input)

	if !strings.Contains(output, "type Handler func(string) error") {
		t.Errorf("expected type alias, got:\n%s", output)
	}
}

func TestGenerateStructWithJsonTag(t *testing.T) {
	input := `type Repo
    Name string as "name"
    Stars int as "stargazers_count"
    URL string as "html_url"
`

	output := generateSource(t, input)

	if !strings.Contains(output, "Name string `json:\"name\"`") {
		t.Errorf("expected Name json tag, got:\n%s", output)
	}
	if !strings.Contains(output, "Stars int `json:\"stargazers_count\"`") {
		t.Errorf("expected Stars json tag, got:\n%s", output)
	}
	if !strings.Contains(output, "URL string `json:\"html_url\"`") {
		t.Errorf("expected URL json tag, got:\n%s", output)
	}
}
