package formatter

import (
	"testing"
)

func TestFormatSimple(t *testing.T) {
	source := `import "fmt"

func main()
    fmt.Println("Hello")
`

	opts := DefaultOptions()
	result, err := Format(source, "test.kuki", opts)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	t.Logf("Result:\n%s", result)
}

func TestFormatWithComments(t *testing.T) {
	source := `# This is a comment
import "fmt"

# Main function
func main()
    # Print hello
    fmt.Println("Hello")
`

	opts := DefaultOptions()
	result, err := Format(source, "test.kuki", opts)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	t.Logf("Result:\n%s", result)
}

func TestFormatGoStyle(t *testing.T) {
	source := `import "fmt"

func main() {
    fmt.Println("Hello")
}
`

	opts := DefaultOptions()
	result, err := Format(source, "test.kuki", opts)
	if err != nil {
		t.Fatalf("Format error: %v", err)
	}

	t.Logf("Result:\n%s", result)
}
