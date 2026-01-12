package changelog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVersionLine(t *testing.T) {
	// Test the version line regex
	if !versionRE.MatchString("# [v1.0.0]") {
		t.Error("versionRE failed to match")
	}
}

func TestParseBulletOverride(t *testing.T) {
	line := "- added override [[PR]](https://github.com/OffchainLabs/prysm/pull/1)"
	res := parseBullet(line, "NOPE")
	if line != res {
		t.Error("parseBullet did not recognize the override")
	}
}

func TestSectionRe(t *testing.T) {
	cases := []struct {
		value string
		valid bool
	}{
		{value: "####### Fixes", valid: false}, // there's no H7
		{value: " Fixes", valid: false},
		{value: "# Fixes", valid: true},
		{value: "## Fixes", valid: true},
		{value: "### Fixes", valid: true},
		{value: "#### Fixes", valid: true},
		{value: "##### Fixes", valid: true},
		{value: "###### Fixes", valid: true},
	}
	for _, c := range cases {
		if valid := sectionRE.MatchString(c.value); valid != c.valid {
			t.Errorf("sectionRE failed to match %v", c)
		}
	}
}

func TestParseFragment_CustomSections(t *testing.T) {
	defaultMap := make(map[string]bool)
	for _, s := range Sections {
		defaultMap[s] = true
	}

	customMap := make(map[string]bool)
	for _, s := range Sections {
		customMap[s] = true
	}
	customMap["Configuration"] = true
	customMap["SpecialFeature"] = true

	tests := []struct {
		name          string
		lines         []string
		validSections map[string]bool
		expectKey     string
		expectCount   int
	}{
		{
			name:          "Standard Fixed Section (Default)",
			lines:         []string{"### Fixed", "- bug fix"},
			validSections: defaultMap,
			expectKey:     "Fixed",
			expectCount:   1,
		},
		{
			name:          "Configuration Section (Returns in parse, fails in validate)",
			lines:         []string{"### Configuration", "- added flag"},
			validSections: defaultMap,
			expectKey:     "Configuration",
			expectCount:   1,
		},
		{
			name:          "Configuration Section (Custom - Should Pass)",
			lines:         []string{"### Configuration", "- added flag"},
			validSections: customMap,
			expectKey:     "Configuration",
			expectCount:   1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			res := ParseFragment(tc.lines, "PR-LINK")
			if len(res[tc.expectKey]) != tc.expectCount {
				t.Errorf("expected %d entries for section '%s', got %d", tc.expectCount, tc.expectKey, len(res[tc.expectKey]))
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	changelogDir := filepath.Join(tmpDir, "changelog")
	err := os.Mkdir(changelogDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	configContent := []byte("sections:\n  - Added\n  - CustomOne\n  - Configuration")
	configPath := filepath.Join(changelogDir, ".unclog.yaml")
	err = os.WriteFile(configPath, configContent, 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig(tmpDir)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}
	if len(cfg.Sections) != 3 {
		t.Errorf("expected 3 sections, got %d", len(cfg.Sections))
	}
}

func TestValidSections_Logic(t *testing.T) {
	allowed := map[string]bool{"Added": true, "Fixed": true}

	tests := []struct {
		name      string
		fragments map[string][]string
		expectErr bool
	}{
		{
			name:      "Valid Sections Only",
			fragments: map[string][]string{"Added": {"- item"}},
			expectErr: false,
		},
		{
			name:      "Contains Ignored (Pass)",
			fragments: map[string][]string{"Ignored": {"- reason"}},
			expectErr: false,
		},
		{
			name:      "Contains Invalid Section (Fail)",
			fragments: map[string][]string{"Removed": {"- item"}},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidSections(tc.fragments, allowed)
			if tc.expectErr && err == nil {
				t.Error("expected validation error, got nil")
			}
			if !tc.expectErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if tc.expectErr && err != nil {
				if !strings.Contains(err.Error(), "Must be one of: Added, Fixed") {
					t.Errorf("Error message missing suggestions: %v", err)
				}
			}
		})
	}
}

func TestSectionHeader_Variations(t *testing.T) {
	tests := []struct {
		line      string
		expectSec string
	}{
		{"# Fixed", "Fixed"},
		{"## Fixed", "Fixed"},
		{"####### Fixed", ""},
		{"### Fixed ", "Fixed"},
		{"### fixed", "fixed"},
	}

	for _, tc := range tests {
		t.Run(tc.line, func(t *testing.T) {
			res := parseSection(tc.line)
			if res != tc.expectSec {
				t.Errorf("input '%s': expected '%s', got '%s'", tc.line, tc.expectSec, res)
			}
		})
	}
}
