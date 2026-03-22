package version

import (
	"regexp"
	"strings"
	"testing"
)

func TestVersionShape(t *testing.T) {
	t.Parallel()

	checks := []struct {
		name string
		run  func(*testing.T)
	}{
		{
			name: "non-empty",
			run: func(t *testing.T) {
				if Version == "" {
					t.Fatal("Version must not be empty")
				}
			},
		},
		{
			name: "no-v-prefix",
			run: func(t *testing.T) {
				if strings.HasPrefix(Version, "v") {
					t.Fatalf("Version should not include a leading v: %q", Version)
				}
			},
		},
		{
			name: "semver-triplet",
			run: func(t *testing.T) {
				semverTriplet := regexp.MustCompile(`^\d+\.\d+\.\d+$`)
				if !semverTriplet.MatchString(Version) {
					t.Fatalf("Version should be a bare semver triplet, got %q", Version)
				}
			},
		},
	}

	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			t.Parallel()
			check.run(t)
		})
	}
}
