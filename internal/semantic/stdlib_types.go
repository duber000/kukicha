package semantic

// goStdlibType holds the TypeKind and optional name for one return position.
// Used by both the Go stdlib and Kukicha stdlib registries.
type goStdlibType struct {
	Kind TypeKind
	Name string // non-empty for TypeKindNamed (e.g. "error")
}

// goStdlibEntry holds the return signature info for a stdlib function:
// return count, per-position type info, and optional parameter names.
// Used by both the Go stdlib and Kukicha stdlib registries.
type goStdlibEntry struct {
	Count      int
	Types      []goStdlibType
	ParamNames []string // Parameter names (populated for Kukicha stdlib; nil for Go stdlib)
}
