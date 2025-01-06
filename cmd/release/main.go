package release

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/OffchainLabs/unclog/changelog"
)

func parseArgs(args []string) (*changelog.Config, error) {
	flags := flag.NewFlagSet("release", flag.ContinueOnError)
	c := &changelog.Config{RepoConfig: changelog.RepoConfig{Owner: "prysmaticlabs", Repo: "prysm"}, ReleaseTime: time.Now()}
	flags.StringVar(&c.RepoPath, "repo", "", "Path to the git repository")
	flags.StringVar(&c.ChangesDir, "changelog-dir", "changelog", "Path to the directory containing changelog fragments for each commit")
	flags.StringVar(&c.Tag, "tag", "", "New release tag (must already exist in repo)")
	flags.StringVar(&c.PreviousPath, "prev", "CHANGELOG.md", "Path to current changelog in the repo. This will be pulled from HEAD")
	flags.BoolVar(&c.Cleanup, "cleanup", false, "Remove the changelog fragment files after generating the changelog")
	flags.Parse(args)
	if c.RepoPath == "" {
		return c, fmt.Errorf("repo is required")
	}
	if c.Tag == "" {
		return c, fmt.Errorf("tag is required")
	}
	if c.PreviousPath == "" {
		return c, fmt.Errorf("prev is required")
	}
	return c, nil
}

func Run(ctx context.Context, args []string) error {
	cfg, err := parseArgs(args)
	if err != nil {
		return err
	}
	out, err := changelog.Release(ctx, cfg)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
