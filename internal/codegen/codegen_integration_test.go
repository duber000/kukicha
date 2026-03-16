package codegen

import (
	"go/parser"
	"go/token"
	"testing"

	kukiparser "github.com/duber000/kukicha/internal/parser"
	"github.com/duber000/kukicha/internal/semantic"
)

// fullPipeline runs the complete Kukicha compilation pipeline:
// lex → parse → semantic → codegen, returning the generated Go source.
// This catches issues that unit tests miss by verifying all stages cooperate.
func fullPipeline(t *testing.T, source, filename string) string {
	t.Helper()

	p, err := kukiparser.New(source, filename)
	if err != nil {
		t.Fatalf("lexer/parser init error: %v", err)
	}

	program, parseErrors := p.Parse()
	if len(parseErrors) > 0 {
		t.Fatalf("parse errors: %v", parseErrors)
	}

	analyzer := semantic.NewWithFile(program, filename)
	semanticErrors := analyzer.Analyze()
	if len(semanticErrors) > 0 {
		t.Fatalf("semantic errors: %v", semanticErrors)
	}

	gen := New(program)
	gen.SetSourceFile(filename)
	gen.SetExprReturnCounts(analyzer.ReturnCounts())
	gen.SetExprTypes(analyzer.ExprTypes())

	output, err := gen.Generate()
	if err != nil {
		t.Fatalf("codegen error: %v", err)
	}

	return output
}

// assertValidGo parses the generated Go source to verify it is syntactically valid.
func assertValidGo(t *testing.T, source string) {
	t.Helper()
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "generated.go", source, parser.AllErrors)
	if err != nil {
		t.Errorf("generated Go is not valid syntax:\n%v\n\nGenerated source:\n%s", err, source)
	}
}

func TestIntegration_SimpleFunction(t *testing.T) {
	source := `func Add(a int, b int) int
    return a + b
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_TypeAndMethod(t *testing.T) {
	source := `type User
    name string
    age int

func GetName on u User string
    return u.name

func SetAge on u reference User
    u.age = 42
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_StringInterpolation(t *testing.T) {
	source := `func greet(name string) string
    return "Hello, {name}!"

func main()
    msg := greet("world")
    print(msg)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_ErrorHandling(t *testing.T) {
	source := `import "os"

func readConfig(path string) (list of byte, error)
    data, err := os.ReadFile(path)
    if err not equals empty
        return empty, error "failed to read config: {err}"
    return data, empty
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_OnErrReturn(t *testing.T) {
	source := `import "os"

func readFile(path string) (list of byte, error)
    data := os.ReadFile(path) onerr return empty, error "{error}"
    return data, empty
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_OnErrDefault(t *testing.T) {
	source := `import "strconv"

func safeAtoi(s string) int
    n := strconv.Atoi(s) onerr 0
    return n
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_OnErrPanic(t *testing.T) {
	source := `import "os"

func mustRead(path string) list of byte
    data := os.ReadFile(path) onerr panic "failed: {error}"
    return data
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_ForRangeLoop(t *testing.T) {
	source := `func sumList(items list of int) int
    total := 0
    for item in items
        total = total + item
    return total
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_ForNumericLoop(t *testing.T) {
	source := `func countUp(n int) int
    total := 0
    for i from 0 to n
        total = total + i
    return total
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_SwitchStatement(t *testing.T) {
	source := `func classify(n int) string
    switch
        when n < 0
            return "negative"
        when n equals 0
            return "zero"
        otherwise
            return "positive"
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_SwitchOnValue(t *testing.T) {
	source := `func describe(cmd string) string
    switch cmd
        when "start", "run"
            return "execute"
        when "stop"
            return "halt"
        otherwise
            return "unknown"
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_ListAndMapLiterals(t *testing.T) {
	source := `func main()
    items := list of string{"hello", "world"}
    config := map of string to int{"port": 8080, "timeout": 30}
    print(items)
    print(config)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_Interface(t *testing.T) {
	source := `interface Greeter
    Greet(name string) string

type FormalGreeter
    prefix string

func Greet on g FormalGreeter string
    return "greetings"
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_GlobalVar(t *testing.T) {
	source := `var AppName string = "myapp"
var Version = "1.0.0"

func main()
    print(AppName)
    print(Version)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_Variadic(t *testing.T) {
	source := `func Sum(many numbers int) int
    total := 0
    for n in numbers
        total = total + n
    return total

func main()
    result := Sum(1, 2, 3)
    print(result)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_Channel(t *testing.T) {
	source := `func main()
    ch := make(channel of string)
    go
        send "hello" to ch
    msg := receive from ch
    print(msg)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_DefaultParams(t *testing.T) {
	source := `func Greet(name string, greeting string = "Hello") string
    return "{greeting}, {name}!"

func main()
    a := Greet("Alice")
    b := Greet("Bob", "Hi")
    print(a)
    print(b)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_TypeAlias(t *testing.T) {
	source := `type Handler func(string) error
type Transform func(int) (string, error)

func main()
    x := 1
    print(x)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_NestedControlFlow(t *testing.T) {
	source := `func process(items list of int) int
    count := 0
    for item in items
        if item > 0
            switch
                when item > 100
                    count = count + 2
                otherwise
                    count = count + 1
    return count
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_MultipleReturnTypes(t *testing.T) {
	source := `func divide(a int, b int) (int, error)
    if b equals 0
        return 0, error "division by zero"
    return a / b, empty
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_NegativeIndex(t *testing.T) {
	source := `func last(items list of string) string
    return items[-1]
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_ArrowLambda(t *testing.T) {
	source := `func apply(f func(int) int, x int) int
    return f(x)

func main()
    result := apply((n int) => n * 2, 5)
    print(result)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_StructWithJsonTags(t *testing.T) {
	source := `type Repo
    Name string as "name"
    Stars int as "stargazers_count"
    URL string as "html_url"
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_DeferStatement(t *testing.T) {
	source := `import "os"

func writeFile(path string) error
    f, err := os.Create(path)
    if err not equals empty
        return err
    defer f.Close()
    return empty
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}

func TestIntegration_ForThrough(t *testing.T) {
	source := `func countdown(n int)
    for i from n through 0
        print(i)
`
	output := fullPipeline(t, source, "test.kuki")
	assertValidGo(t, output)
}
