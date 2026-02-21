package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	kukicha "github.com/duber000/kukicha"
	"github.com/duber000/kukicha/internal/version"
	"golang.org/x/mod/modfile"
)

const stdlibDirName = ".kukicha/stdlib"
const stdlibVersionFile = "KUKICHA_VERSION"

// stdlibGoMod is the go.mod content for the extracted stdlib module.
// This declares the stdlib as a standalone Go module so user projects can
// reference it via a replace directive.
const stdlibGoMod = `module github.com/duber000/kukicha/stdlib

go 1.25

require (
	github.com/a2aproject/a2a-go v0.3.6
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250715232539-7130f93afb79 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250715232539-7130f93afb79 // indirect
	google.golang.org/grpc v1.73.0 // indirect
	google.golang.org/protobuf v1.36.6 // indirect
)
`

// stdlibGoSum contains dependency checksums for the stdlib module.
const stdlibGoSum = `github.com/a2aproject/a2a-go v0.3.6 h1:VbRoM2MNsfc7o4GkjGt3KZCjbqILAJq846K1z8rpHTc=
github.com/a2aproject/a2a-go v0.3.6/go.mod h1:I7Cm+a1oL+UT6zMoP+roaRE5vdfUa1iQGVN8aSOuZ0I=
github.com/go-logr/logr v1.4.2 h1:6pFjapn8bFcIbiKo3XT4j/BhANplGihG6tvd+8rYgrY=
github.com/go-logr/logr v1.4.2/go.mod h1:9T104GzyrTigFIr8wt5mBrctHMim0Nb2HLGrmQ40KvY=
github.com/go-logr/stdr v1.2.2 h1:hSWxHoqTgW2S2qGc0LTAI563KZ5YKYRhT3MFKZMbjag=
github.com/go-logr/stdr v1.2.2/go.mod h1:mMo/vtBO5dYbehREoey6XUKy/eSumjCCveDpRre4VKE=
github.com/golang/protobuf v1.5.4 h1:i7eJL8qZTpSEXOPTxNKhASYpMn+8e5Q6AdndVa1dWek=
github.com/golang/protobuf v1.5.4/go.mod h1:lnTiLA8Wa4RWRcIUkrtSVa5nRhsEGBg48fD6rSs7xps=
github.com/google/go-cmp v0.7.0 h1:wk8382ETsv4JYUZwIsn6YpYiWiBsYLSJiTsyBybVuN8=
github.com/google/go-cmp v0.7.0/go.mod h1:pXiqmnSA92OHEEa9HXL2W4E7lf9JzCmGVUdgjX3N/iU=
github.com/google/uuid v1.6.0 h1:NIvaJDMOsjHA8n1jAhLSgzrAzy1Hgr+hNrb57e+94F0=
github.com/google/uuid v1.6.0/go.mod h1:TIyPZe4MgqvfeYDBFedMoGGpEw/LqOeaOT+nhxU+yHo=
go.opentelemetry.io/auto/sdk v1.1.0 h1:cH53jehLUN6UFLY71z+NDOiNJqDdPRaXzTel0sJySYA=
go.opentelemetry.io/auto/sdk v1.1.0/go.mod h1:3wSPjt5PWp2RhlCcmmOial7AvC4DQqZb7a7wCow3W8A=
go.opentelemetry.io/otel v1.35.0 h1:xKWKPxrxB6OtMCbmMY021CqC45J+3Onta9MqjhnusiQ=
go.opentelemetry.io/otel v1.35.0/go.mod h1:UEqy8Zp11hpkUrL73gSlELM0DupHoiq72dR+Zqel/+Y=
go.opentelemetry.io/otel/metric v1.35.0 h1:0znxYu2SNyuMSQT4Y9WDWej0VpcsxkuklLa4/siN90M=
go.opentelemetry.io/otel/metric v1.35.0/go.mod h1:nKVFgxBZ2fReX6IlyW28MgZojkoAkJGaE8CpgeAU3oE=
go.opentelemetry.io/otel/sdk v1.35.0 h1:iPctf8iprVySXSKJffSS79eOjl9pvxV9ZqOWT0QejKY=
go.opentelemetry.io/otel/sdk v1.35.0/go.mod h1:+ga1bZliga3DxJ3CQGg3updiaAJoNECOgJREo9KHGQg=
go.opentelemetry.io/otel/sdk/metric v1.35.0 h1:1RriWBmCKgkeHEhM7a2uMjMUfP7MsOF5JpUCaEqEI9o=
go.opentelemetry.io/otel/sdk/metric v1.35.0/go.mod h1:is6XYCUMpcKi+ZsOvfluY5YstFnhW0BidkR+gL+qN+w=
go.opentelemetry.io/otel/trace v1.35.0 h1:dPpEfJu1sDIqruz7BHFG3c7528f6ddfSWfFDVt/xgMs=
go.opentelemetry.io/otel/trace v1.35.0/go.mod h1:WUk7DtFp1Aw2MkvqGdwiXYDZZNvA/1J8o6xRXLrIkyc=
golang.org/x/net v0.41.0 h1:vBTly1HeNPEn3wtREYfy4GZ/NECgw2Cnl+nK6Nz3uvw=
golang.org/x/net v0.41.0/go.mod h1:B/K4NNqkfmg07DQYrbwvSluqCJOOXwUjeb/5lOisjbA=
golang.org/x/sync v0.15.0 h1:KWH3jNZsfyT6xfAfKiz6MRNmd46ByHDYaZ7KSkCtdW8=
golang.org/x/sync v0.15.0/go.mod h1:1dzgHSNfp02xaA81J2MS99Qcpr2w7fw1gpm99rleRqA=
golang.org/x/sys v0.33.0 h1:q3i8TbbEz+JRD9ywIRlyRAQbM0qF7hu24q3teo2hbuw=
golang.org/x/sys v0.33.0/go.mod h1:BJP2sWEmIv4KK5OTEluFJCKSidICx8ciO85XgH3Ak8k=
golang.org/x/text v0.26.0 h1:P42AVeLghgTYr4+xUnTRKDMqpar+PtX7KWuNQL21L8M=
golang.org/x/text v0.26.0/go.mod h1:QK15LZJUUQVJxhz7wXgxSy/CJaTFjd0G+YLonydOVQA=
google.golang.org/genproto/googleapis/api v0.0.0-20250715232539-7130f93afb79 h1:iOye66xuaAK0WnkPuhQPUFy8eJcmwUXqGGP3om6IxX8=
google.golang.org/genproto/googleapis/api v0.0.0-20250715232539-7130f93afb79/go.mod h1:HKJDgKsFUnv5VAGeQjz8kxcgDP0HoE0iZNp0OdZNlhE=
google.golang.org/genproto/googleapis/rpc v0.0.0-20250715232539-7130f93afb79 h1:1ZwqphdOdWYXsUHgMpU/101nCtf/kSp9hOrcvFsnl10=
google.golang.org/genproto/googleapis/rpc v0.0.0-20250715232539-7130f93afb79/go.mod h1:qQ0YXyHHx3XkvlzUtpXDkS29lDSafHMZBAZDc03LQ3A=
google.golang.org/grpc v1.73.0 h1:VIWSmpI2MegBtTuFt5/JWy2oXxtjJ/e89Z70ImfD2ok=
google.golang.org/grpc v1.73.0/go.mod h1:50sbHOUqWoCQGI8V2HQLJM0B+LMlIUjNSZmow7EVBQc=
google.golang.org/protobuf v1.36.6 h1:z1NpPI8ku2WgiWnf+t9wTPsn6eP1L7ksHUlkfLvd9xY=
google.golang.org/protobuf v1.36.6/go.mod h1:jduwjTPXsFjZGTmRluh+L6NjiWu7pchiJ2/5YcXBHnY=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405 h1:yhCVgyC4o1eVCa2tZl7eS0r+SDo693bJlVdllGtEeKM=
gopkg.in/check.v1 v0.0.0-20161208181325-20d25e280405/go.mod h1:Co6ibVJAznAaIkqp8huTwlJQCZ016jof/cbN4VW5Yz0=
gopkg.in/yaml.v3 v3.0.1 h1:fxVm/GzAzEWqLHuvctI91KS9hhNmmWOoWu0XTYJS7CA=
gopkg.in/yaml.v3 v3.0.1/go.mod h1:K4uyk7z7BCEPqu6E+C64Yfv1cQ7kz7rIZviUmN+EgEM=
`

// ensureStdlib extracts the embedded stdlib to projectDir/.kukicha/stdlib/ if not present
// or if the cached version stamp doesn't match the current binary version.
// Returns the absolute path to the extracted stdlib directory.
func ensureStdlib(projectDir string) (string, error) {
	stdlibPath := filepath.Join(projectDir, stdlibDirName)

	// Check version stamp: only skip extraction if cache exists AND matches current version.
	stampPath := filepath.Join(stdlibPath, stdlibVersionFile)
	if stamp, err := os.ReadFile(stampPath); err == nil {
		if strings.TrimSpace(string(stamp)) == version.Version {
			return stdlibPath, nil
		}
		// Version mismatch — remove stale cache and re-extract.
		_ = os.RemoveAll(stdlibPath)
	}

	// Extract from embedded FS
	if err := extractStdlib(stdlibPath); err != nil {
		return "", fmt.Errorf("extracting stdlib: %w", err)
	}

	// Extract agent docs alongside stdlib (tied to the same version stamp).
	if err := extractAgentDocs(projectDir); err != nil {
		return "", fmt.Errorf("extracting agent docs: %w", err)
	}

	return stdlibPath, nil
}

// extractStdlib writes the embedded stdlib files to the target directory,
// plus a generated go.mod and go.sum for the standalone module.
func extractStdlib(targetDir string) error {
	// Walk embedded FS and extract all files under "stdlib/"
	err := fs.WalkDir(kukicha.StdlibFS, "stdlib", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Map embedded path "stdlib/json/json.go" -> targetDir + "/json/json.go"
		relPath, _ := filepath.Rel("stdlib", path)
		targetPath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		data, readErr := kukicha.StdlibFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		return os.WriteFile(targetPath, data, 0644)
	})
	if err != nil {
		return err
	}

	// Write the generated go.mod for the extracted stdlib module
	if err := os.WriteFile(filepath.Join(targetDir, "go.mod"), []byte(stdlibGoMod), 0644); err != nil {
		return err
	}

	// Write the go.sum
	if err := os.WriteFile(filepath.Join(targetDir, "go.sum"), []byte(stdlibGoSum), 0644); err != nil {
		return err
	}

	// Write the version stamp so future runs can detect stale caches.
	if err := os.WriteFile(filepath.Join(targetDir, stdlibVersionFile), []byte(version.Version), 0644); err != nil {
		return err
	}

	return nil
}

const skillStart = "<!-- kukicha:start -->"
const skillEnd = "<!-- kukicha:end -->"

// extractAgentDocs upserts the Kukicha skill section into AGENTS.md in the
// user's project, and appends @AGENTS.md to CLAUDE.md if present.
// Both operations are idempotent. Called from ensureStdlib; shares the
// KUKICHA_VERSION stamp so docs stay in sync with the stdlib.
func extractAgentDocs(projectDir string) error {
	content, err := kukicha.SkillFS.ReadFile("docs/SKILL.md")
	if err != nil {
		return fmt.Errorf("reading embedded docs/SKILL.md: %w", err)
	}

	if err := upsertSkillSection(filepath.Join(projectDir, "AGENTS.md"), string(content)); err != nil {
		return fmt.Errorf("updating AGENTS.md: %w", err)
	}

	claudePath := filepath.Join(projectDir, "CLAUDE.md")
	if _, err := os.Stat(claudePath); err == nil {
		if err := appendIfMissing(claudePath, "@AGENTS.md"); err != nil {
			return fmt.Errorf("updating CLAUDE.md: %w", err)
		}
	}

	return nil
}

// upsertSkillSection inserts or replaces the kukicha skill block in the file
// at path. The block is delimited by HTML comments so it can be updated in
// place across `kukicha init` runs without duplicating content.
func upsertSkillSection(path, content string) error {
	section := skillStart + "\n" + content + "\n" + skillEnd + "\n"

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.WriteFile(path, []byte(section), 0644)
		}
		return err
	}

	s := string(data)
	startIdx := strings.Index(s, skillStart)
	endIdx := strings.Index(s, skillEnd)

	if startIdx == -1 || endIdx == -1 || endIdx < startIdx {
		// Section not present — append it.
		if !strings.HasSuffix(s, "\n") {
			s += "\n"
		}
		return os.WriteFile(path, []byte(s+"\n"+section), 0644)
	}

	// Replace existing section in place.
	after := strings.TrimPrefix(s[endIdx+len(skillEnd):], "\n")
	return os.WriteFile(path, []byte(s[:startIdx]+section+after), 0644)
}

// appendIfMissing appends line to the file at path if it is not already present.
func appendIfMissing(path, line string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), line) {
		return nil
	}
	s := string(data)
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	return os.WriteFile(path, []byte(s+line+"\n"), 0644)
}

// ensureGoMod checks the project's go.mod and adds the stdlib require/replace
// directives if they are not already present.
func ensureGoMod(projectDir, stdlibPath string) error {
	goModPath := filepath.Join(projectDir, "go.mod")

	data, err := os.ReadFile(goModPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no go.mod found in %s; run 'go mod init <module>' or 'kukicha init' first", projectDir)
		}
		return err
	}

	mod, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}

	// Calculate relative path from project dir to stdlib
	relStdlib, err := filepath.Rel(projectDir, stdlibPath)
	if err != nil {
		relStdlib = stdlibPath
	}

	const stdlibModule = "github.com/duber000/kukicha/stdlib"
	const stdlibVersion = "v0.0.0"

	// Add require if missing
	if !hasRequire(mod, stdlibModule) {
		if err := mod.AddRequire(stdlibModule, stdlibVersion); err != nil {
			return fmt.Errorf("adding require: %w", err)
		}
	}

	// Add or update replace
	relPath := "./" + filepath.ToSlash(relStdlib)
	if err := mod.AddReplace(stdlibModule, "", relPath, ""); err != nil {
		return fmt.Errorf("adding replace: %w", err)
	}

	formatted, err := mod.Format()
	if err != nil {
		return fmt.Errorf("formatting go.mod: %w", err)
	}

	return os.WriteFile(goModPath, formatted, 0644)
}

// needsStdlib checks if the generated Go code imports any Kukicha stdlib packages.
// Returns false if we're inside the kukicha repo itself (stdlib already available).
func needsStdlib(goCode string) bool {
	if !strings.Contains(goCode, "github.com/duber000/kukicha/stdlib/") {
		return false
	}
	// Don't extract stdlib if we're inside the kukicha repo itself
	if isKukichaRepo() {
		return false
	}

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", goCode, parser.ImportsOnly)
	if err != nil {
		// Fallback to substring check if parsing fails
		return true
	}

	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, "\"")
		if strings.HasPrefix(path, "github.com/duber000/kukicha/stdlib/") {
			return true
		}
	}
	return false
}

func hasRequire(mod *modfile.File, path string) bool {
	for _, req := range mod.Require {
		if req.Mod.Path == path {
			return true
		}
	}
	return false
}

// isKukichaRepo checks if the current working directory is inside the kukicha repo.
// This is detected by checking if go.mod declares module github.com/duber000/kukicha.
func isKukichaRepo() bool {
	cwd, err := os.Getwd()
	if err != nil {
		return false
	}
	// Walk up looking for go.mod
	for d := cwd; d != filepath.Dir(d); d = filepath.Dir(d) {
		goModPath := filepath.Join(d, "go.mod")
		data, err := os.ReadFile(goModPath)
		if err != nil {
			continue
		}
		// Check if this is the kukicha repo's go.mod
		content := string(data)
		if strings.Contains(content, "module github.com/duber000/kukicha\n") ||
			strings.Contains(content, "module github.com/duber000/kukicha\r\n") {
			return true
		}
		// Found a go.mod but it's not the kukicha repo
		return false
	}
	return false
}

// findProjectDir walks up from the given filename to find the directory
// containing a go.mod file. If none is found, returns the directory of the file.
func findProjectDir(filename string) string {
	dir := filepath.Dir(filename)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return dir
	}

	// Walk up looking for go.mod
	for d := absDir; d != filepath.Dir(d); d = filepath.Dir(d) {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
	}

	return absDir
}
