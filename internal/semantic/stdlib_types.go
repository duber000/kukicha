package semantic

// goStdlibType holds the TypeKind and optional name for one return position.
// Used by both the Go stdlib and Kukicha stdlib registries.
type goStdlibType struct {
	Kind        TypeKind
	Name        string         // non-empty for TypeKindNamed (e.g. "error")
	ElementType *goStdlibType  // element type for TypeKindList (e.g. list of string → &goStdlibType{Kind: TypeKindString})
	KeyType     *goStdlibType  // key type for TypeKindMap
	ValueType   *goStdlibType  // value type for TypeKindMap
}

// goStdlibEntry holds the return signature info for a stdlib function:
// return count, per-position type info, and optional parameter names.
// Used by both the Go stdlib and Kukicha stdlib registries.
type goStdlibEntry struct {
	Count           int
	Types           []goStdlibType
	ParamNames      []string         // Parameter names (populated for Kukicha stdlib; nil for Go stdlib)
	DefaultValues   []string         // Go expression strings for default parameter values; "" = no default
	ParamFuncParams map[int][]goStdlibType // func-typed param index → inner param types (for lambda inference)
}

// GetStdlibEntry returns the Kukicha stdlib registry entry for the given
// qualified name (e.g., "string.PadRight"). Returns the entry and true if found.
func GetStdlibEntry(name string) (goStdlibEntry, bool) {
	entry, ok := generatedStdlibRegistry[name]
	return entry, ok
}

// GetSliceGenericClass returns the generic classification for a stdlib/slice
// function: "T" (uses any), "K" (uses any2), "TK" (uses both), or "" (not generic).
func GetSliceGenericClass(qualifiedName string) string {
	return generatedSliceGenericClass[qualifiedName]
}

// GetSecurityCategory returns the security check category for a stdlib function
// (e.g., "sql", "html", "fetch", "files", "redirect", "shell"), or "" if none.
func GetSecurityCategory(qualifiedName string) string {
	return generatedSecurityFunctions[qualifiedName]
}

// IsKnownInterface returns true if the qualified type name is a known interface
// from either the Go stdlib or the Kukicha stdlib registries.
func IsKnownInterface(qualifiedName string) bool {
	return generatedGoInterfaces[qualifiedName] || generatedStdlibInterfaces[qualifiedName]
}
