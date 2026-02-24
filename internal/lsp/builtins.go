package lsp

// BuiltinInfo holds metadata for a builtin function or keyword-level construct.
type BuiltinInfo struct {
	Name      string // e.g. "print"
	Signature string // e.g. "func print(args ...any)"
	Doc       string // e.g. "Prints values to stdout"
}

// builtins is the single source of truth for builtin function metadata
// used by both completion and hover handlers.
var builtins = []BuiltinInfo{
	{"print", "func print(args ...any)", "Prints values to stdout"},
	{"len", "func len(v any) int", "Returns the length of a string, list, or map"},
	{"append", "func append(slice []T, elems ...T) []T", "Appends elements to a slice"},
	{"make", "func make(T type, size ...int) T", "Creates a slice, map, or channel"},
	{"min", "func min(x T, y T, rest ...T) T", "Returns the minimum of its arguments (Go 1.21+)"},
	{"max", "func max(x T, y T, rest ...T) T", "Returns the maximum of its arguments (Go 1.21+)"},
	{"close", "func close(ch chan T)", "Closes a channel"},
	{"panic", "func panic(v any)", "Stops normal execution and begins panicking"},
	{"recover", "func recover() any", "Regains control of a panicking goroutine"},
	{"empty", "empty T", "Returns the zero value of type T"},
	{"error", "error \"message\"", "Creates a new error with the given message"},
}

// builtinCompletions returns all builtin entries for use by completion.
func builtinCompletions() []BuiltinInfo {
	return builtins
}

// lookupBuiltin returns "signature\ndoc" for the named builtin, or "" if not found.
func lookupBuiltin(name string) string {
	for _, b := range builtins {
		if b.Name == name {
			return b.Signature + "\n" + b.Doc
		}
	}
	return ""
}
