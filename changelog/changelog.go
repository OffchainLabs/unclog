package changelog

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
)

var versionRE = regexp.MustCompile(`^#+ \[(v\d+\.\d+\.\d+)\]`)
var sectionRE = regexp.MustCompile(`^### (\w+)\s?$`)

const preamble = `# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.`

// Sections represents all the possible sections in the changelog, in the desired order.
var Sections = []string{"Added", "Changed", "Deprecated", "Removed", "Fixed", "Security"}

var sectionIgnored = "Ignored"

// map lookup for valid section names, populated from sectionOrder in init()
var sectionNames = make(map[string]bool)

type RepoConfig struct {
	Owner   string `json:"owner"`
	Repo    string `json:"repo"`
	MainRev string `json:"main_rev"`
}

type Config struct {
	RepoPath     string
	Repository   *git.Repository
	ChangesDir   string
	Tag          string
	PreviousPath string
	Previous     Previous
	RepoConfig   RepoConfig
	Cleanup      bool
	Branch       string
	ReleaseTime  time.Time
	OutputPath   string
}

func (c *Config) Repo() (*git.Repository, error) {
	if c.Repository != nil {
		return c.Repository, nil
	}
	r, err := git.PlainOpen(c.RepoPath)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func versionFromLine(line string) string {
	if !versionRE.MatchString(line) {
		return ""
	}
	return versionRE.FindStringSubmatch(line)[1]
}

type Previous struct {
	Version string
	Body    string
}

// ParsePreviousVersionBody drops the changelog preamble and anything up to the previous release
// header containing the semver of the last release. It returns the version and the rest of the body
// including the previous header. These values can be used together with the hardcoded preamble and
// the new release changelog to assemble the final combined changelog.
func NewPreviousChangelog(r io.Reader) (Previous, error) {
	p := Previous{}
	scn := bufio.NewScanner(r)
	for scn.Scan() {
		line := scn.Text()
		if p.Version != "" {
			p.Body += "\n" + line
			continue
		}
		v := versionFromLine(line)
		if v != "" {
			p.Version = v
			p.Body += line
		}
	}
	if p.Version == "" {
		return p, fmt.Errorf("no version found")
	}
	return p, nil
}

// Each commit's changelog file can have entries in different sections, indicated by
// the same section headers as the final changelog.
// Merge all changelog bullet points into their respective sections.
func mergeEntries(fragments []Fragment) map[string][]string {
	sections := make(map[string][]string)
	for _, f := range fragments {
		pr := f.Commit.prLink()
		csecs := ParseFragment(f.Lines, pr)
		for k, v := range csecs {
			sections[k] = append(sections[k], v...)
		}
	}
	return sections
	/*
		fragment, err := findFragment(cfg.ChangesDir, cm)
		if err != nil {
			if errors.Is(err, errNoChangelogFragment) {
				log.Printf("no changelog fragment found for commit %s", cm.sha)
				return nil
			}
			return err
		}
		pr := cm.prLink()
		csecs := parseFragments(fragment.lines, pr)
		for k, v := range csecs {
			sections[k] = append(sections[k], v...)
		}
		return nil
	*/
}

func findFragments(dir string, commits []Commit) ([]Fragment, error) {
	fragments := make([]Fragment, 0)
	for _, cm := range commits {
		parent, err := cm.Parent()
		if err != nil {
			return fragments, err
		}
		f, err := FindFragment(dir, parent, cm)
		if err != nil {
			if errors.Is(err, errNoChangelogFragment) {
				log.Printf("no changelog fragment found for commit %s", cm.Id())
				continue
			}
			return nil, err
		}
		fragments = append(fragments, f)
	}
	return fragments, nil
}

func Release(ctx context.Context, cfg *Config) (string, error) {
	prd, pcl, err := getFile(cfg, cfg.PreviousPath)
	if err != nil {
		return "", err
	}
	defer pcl.Close()
	prev, err := NewPreviousChangelog(prd)
	if err != nil {
		return "", err
	}
	cfg.Previous = prev

	commits, err := commitsAfter(cfg)
	if err != nil {
		return "", err
	}
	fragments, err := findFragments(cfg.ChangesDir, commits)
	if err != nil {
		return "", err
	}
	sections := mergeEntries(fragments)
	if cfg.Cleanup {
		if err := cleanupFragments(cfg, fragments); err != nil {
			return "", err
		}
	}
	body := preamble + "\n\n" + header(cfg)
	for _, s := range Sections {
		bs, ok := sections[s]
		if !ok || len(bs) == 0 {
			continue
		}
		body += formatSection(s, bs)
	}
	return body + "\n\n" + cfg.Previous.Body, nil
}

func formatSection(name string, bullets []string) string {
	section := "\n\n### " + name + "\n"
	for _, b := range bullets {
		section += "\n" + b
	}
	return section
}

func header(cfg *Config) string {
	// ## [v5.1.1](https://github.com/prysmaticlabs/prysm/compare/v5.1.0...v5.1.1) - 2024-10-15
	return fmt.Sprintf("## [%s](https://github.com/%s/%s/compare/%s...%s) - %s",
		cfg.Tag,
		cfg.RepoConfig.Owner, cfg.RepoConfig.Repo,
		cfg.Previous.Version, cfg.Tag,
		cfg.ReleaseTime.Format("2006-01-02"),
	)
}

func parseSection(line string) string {
	sec := sectionRE.FindStringSubmatch(line)
	if len(sec) == 0 {
		return ""
	}
	// Special case to allow PRs that do not create an entry in the changelog.
	if sec[1] == sectionIgnored {
		return sectionIgnored
	}
	if _, ok := sectionNames[sec[1]]; !ok {
		return ""
	}
	return sec[1]
}

var prLinkRE = regexp.MustCompile(`\[\[PR\]\]\(https:\/\/[^\)]+\)`)

func parseBullet(line string, pr string) string {
	trimmed := strings.TrimLeft(line, " ")
	if !strings.HasPrefix(trimmed, "- ") {
		return ""
	}
	if prLinkRE.Match([]byte(trimmed)) {
		return line
	}
	return strings.TrimRight(line, " .") + ". " + pr
}

func ParseFragment(lines []string, pr string) map[string][]string {
	fragments := make(map[string][]string)
	var current string
	for _, line := range lines {
		section := parseSection(line)
		if section != "" {
			current = section
			continue
		}
		bullet := parseBullet(line, pr)
		if bullet == "" {
			continue
		}
		fragments[current] = append(fragments[current], bullet)
	}
	return fragments
}

func init() {
	for _, s := range Sections {
		sectionNames[s] = true
	}
}

func ValidSections(sections map[string][]string) error {
	if len(sections) == 0 {
		return errors.New("no changelog sections found")
	}
	for k := range sections {
		if _, ok := sectionNames[k]; !ok {
			if k == sectionIgnored {
				continue
			}
			return fmt.Errorf("invalid section name: %s", k)
		}
	}
	return nil
}
