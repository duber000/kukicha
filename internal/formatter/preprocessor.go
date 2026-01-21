package formatter

import (
	"strings"
	"unicode"
)

// Preprocessor converts Go-style syntax to Kukicha-style indentation
type Preprocessor struct {
	source      []rune
	output      strings.Builder
	current     int
	line        int
	indentLevel int
	indentStr   string
}

// NewPreprocessor creates a new preprocessor
func NewPreprocessor(source string) *Preprocessor {
	return &Preprocessor{
		source:    []rune(source),
		indentStr: "    ", // 4 spaces
	}
}

// Process converts Go-style braces to Kukicha-style indentation
// It handles:
// - Lines ending with { -> remove brace, increase indent for following lines
// - Lines that are just } -> remove brace, decrease indent
// - Trailing semicolons -> remove
// - Struct/map literals are preserved (braces in expressions)
func (p *Preprocessor) Process() string {
	source := string(p.source)

	// Check if the source uses Go-style braces for blocks
	// If not, return the source unchanged (preserving existing indentation)
	if !p.hasGoStyleBraces(source) {
		return source
	}

	lines := strings.Split(source, "\n")
	var result []string

	indentLevel := 0

	for i, line := range lines {
		processed := p.processLine(line, &indentLevel, i, lines)
		if processed != "" || (i < len(lines)-1) { // Keep empty lines except trailing
			result = append(result, processed)
		}
	}

	// Join and ensure single trailing newline
	output := strings.Join(result, "\n")
	output = strings.TrimRight(output, "\n") + "\n"
	return output
}

// hasGoStyleBraces checks if the source uses Go-style braces for blocks
// (not just for struct/map literals)
func (p *Preprocessor) hasGoStyleBraces(source string) bool {
	lines := strings.Split(source, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for lines ending with { that are control flow, not literals
		if strings.HasSuffix(trimmed, "{") && !p.isExpressionBrace(trimmed) {
			return true
		}

		// Check for standalone closing brace (indicates Go-style blocks)
		if trimmed == "}" {
			return true
		}

		// Check for } else { pattern
		if strings.HasPrefix(trimmed, "} else") {
			return true
		}
	}

	return false
}

func (p *Preprocessor) processLine(line string, indentLevel *int, lineIdx int, allLines []string) string {
	trimmed := strings.TrimSpace(line)

	// Skip empty lines
	if trimmed == "" {
		return ""
	}

	// Handle closing brace only line
	if trimmed == "}" {
		*indentLevel--
		if *indentLevel < 0 {
			*indentLevel = 0
		}
		return "" // Remove the brace-only line
	}

	// Handle closing brace with else: } else {
	if strings.HasPrefix(trimmed, "} else") {
		*indentLevel--
		if *indentLevel < 0 {
			*indentLevel = 0
		}
		// Process the else part
		remaining := strings.TrimPrefix(trimmed, "} ")
		return p.processLine(remaining, indentLevel, lineIdx, allLines)
	}

	// Calculate current indentation
	currentIndent := strings.Repeat(p.indentStr, *indentLevel)

	// Remove trailing semicolon
	if strings.HasSuffix(trimmed, ";") {
		trimmed = strings.TrimSuffix(trimmed, ";")
		trimmed = strings.TrimSpace(trimmed)
	}

	// Check if line ends with opening brace (but not in a string or struct literal context)
	if strings.HasSuffix(trimmed, "{") && !p.isExpressionBrace(trimmed) {
		// Remove the trailing brace
		trimmed = strings.TrimSuffix(trimmed, "{")
		trimmed = strings.TrimSpace(trimmed)
		result := currentIndent + trimmed
		*indentLevel++
		return result
	}

	return currentIndent + trimmed
}

// isExpressionBrace determines if a line's trailing brace is part of an expression
// (struct literal, map literal) rather than a block opener
func (p *Preprocessor) isExpressionBrace(line string) bool {
	line = strings.TrimSpace(line)

	// If line contains := or = followed by something ending with {, it's likely a literal
	// e.g., "x := MyStruct{" or "y = map[string]int{"
	if strings.Contains(line, ":=") || (strings.Contains(line, "=") && !strings.Contains(line, "==")) {
		// Check if there's a type before the brace
		beforeBrace := strings.TrimSuffix(line, "{")
		beforeBrace = strings.TrimSpace(beforeBrace)

		// Common literal patterns
		if strings.HasSuffix(beforeBrace, "]") { // slice/array type: []int{
			return true
		}
		if matched := isTypeName(beforeBrace); matched {
			return true
		}
	}

	// Check for struct/map/slice literal patterns
	// TypeName{ or ]Type{ or map[...]Type{
	beforeBrace := strings.TrimSuffix(line, "{")
	beforeBrace = strings.TrimSpace(beforeBrace)

	// If it ends with a type indicator, it's a literal
	if strings.HasSuffix(beforeBrace, "]") {
		return true // []Type{ or [n]Type{
	}

	// Check for map type pattern
	if strings.Contains(beforeBrace, "map[") && strings.Contains(beforeBrace, "]") {
		return true
	}

	// Check for struct literal: ends with a capitalized identifier or package.Type
	words := strings.Fields(beforeBrace)
	if len(words) > 0 {
		lastWord := words[len(words)-1]
		// Remove assignment operators
		lastWord = strings.TrimSuffix(lastWord, ":=")
		lastWord = strings.TrimSuffix(lastWord, "=")
		lastWord = strings.TrimSpace(lastWord)

		// If ends with package.Type or TypeName, it's likely a struct literal
		if strings.Contains(lastWord, ".") {
			parts := strings.Split(lastWord, ".")
			typeName := parts[len(parts)-1]
			if len(typeName) > 0 && unicode.IsUpper(rune(typeName[0])) {
				return true
			}
		} else if len(lastWord) > 0 && unicode.IsUpper(rune(lastWord[0])) {
			// Could be struct literal, but also could be: if Foo{} or similar
			// Check if preceded by control keyword
			controlKeywords := []string{"if", "for", "func", "else", "type", "interface"}
			for _, kw := range controlKeywords {
				if strings.HasPrefix(line, kw+" ") || line == kw {
					return false // It's a control structure, not a literal
				}
			}
			return true
		}
	}

	return false
}

// isTypeName checks if the end of a string looks like a type name for a struct literal
func isTypeName(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return false
	}

	// Get the last word
	words := strings.Fields(s)
	if len(words) == 0 {
		return false
	}

	lastWord := words[len(words)-1]

	// Check for package.Type pattern
	if strings.Contains(lastWord, ".") {
		parts := strings.Split(lastWord, ".")
		typeName := parts[len(parts)-1]
		return len(typeName) > 0 && unicode.IsUpper(rune(typeName[0]))
	}

	// Check for simple TypeName (starts with uppercase)
	return unicode.IsUpper(rune(lastWord[0]))
}

// ProcessSource is a convenience function to preprocess source code
func ProcessSource(source string) string {
	pp := NewPreprocessor(source)
	return pp.Process()
}
