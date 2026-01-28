package kukicha

import "embed"

// StdlibFS contains the embedded Kukicha standard library source files.
// This includes all transpiled .go files from stdlib sub-packages.
// The .kuki source files are not embedded since only the Go code is needed.
// A go.mod file for the extracted stdlib is generated at extraction time.
//
//go:embed stdlib/*/*.go
var StdlibFS embed.FS
