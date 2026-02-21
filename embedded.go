package kukicha

import "embed"

// StdlibFS contains the embedded Kukicha standard library source files.
// This includes all transpiled .go files from stdlib sub-packages.
// The .kuki source files are not embedded since only the Go code is needed.
// A go.mod file for the extracted stdlib is generated at extraction time.
//
//go:embed stdlib/*/*.go
var StdlibFS embed.FS

// SkillFS contains docs/SKILL.md — the concise Kukicha language reference
// for AI coding agents. Extracted and upserted into AGENTS.md in user projects
// by `kukicha init`, tied to the same KUKICHA_VERSION stamp as the stdlib.
//
//go:embed docs/SKILL.md
var SkillFS embed.FS
