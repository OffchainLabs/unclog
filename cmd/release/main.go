package release

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
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
	flags.StringVar(&c.OutputPath, "output", "", "Path to file where merged output will be written (relative to the -repo flag). Defaults to the value of the -prev flag")
	flags.BoolVar(&c.Cleanup, "cleanup", false, "Remove the changelog fragment files after generating the changelog")
	flags.Parse(args)
	if c.RepoPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("repo flag not set and can't get working directory from syscall, %w", err)
		}
		c.RepoPath = wd
	}
	if c.Tag == "" {
		return c, fmt.Errorf("tag is required")
	}
	if c.PreviousPath == "" {
		return c, fmt.Errorf("prev is required")
	}
	if c.OutputPath == "" {
		c.OutputPath = c.PreviousPath
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
	clPath := path.Join(cfg.RepoPath, cfg.OutputPath)
	if err := os.WriteFile(clPath, []byte(out), 0644); err != nil {
		return err
	}
	return nil
}
