package release

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/OffchainLabs/unclog/changelog"
)

func requireNoError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func commitOpts(when time.Time) *git.CommitOptions {
	return &git.CommitOptions{Author: &object.Signature{Name: "test", Email: "a@b.c", When: when}}
}

func copyFileToRepo(t *testing.T, repo *git.Repository, fname string, ctime time.Time, prNum int, tag string) {
	tdp := path.Join("testdata", fname)
	clp := path.Join("changelog", fname)
	fh, err := os.Open(tdp)
	requireNoError(t, err)
	defer fh.Close()
	tree, err := repo.Worktree()
	requireNoError(t, err)
	outfh, err := tree.Filesystem.Create(clp)
	defer outfh.Close()
	_, err = io.Copy(outfh, fh)
	requireNoError(t, err)
	commitAddTag(t, repo, clp, prNum, ctime, tag)
}

func commitAddTag(t *testing.T, repo *git.Repository, fp string, prNum int, ctime time.Time, tag string) {
	tree, err := repo.Worktree()
	requireNoError(t, err)
	_, err = tree.Add(fp)
	requireNoError(t, err)
	msg := fmt.Sprintf("%s (#%d)", fp, prNum)
	ct, err := tree.Commit(msg, commitOpts(ctime))
	if tag != "" {
		_, err = repo.CreateTag(tag, ct, nil)
		requireNoError(t, err)
	}
}

func TestExamples(t *testing.T) {
	repo, cfg, prevTime, prNum := setupTestRepo(t)
	prNum++
	copyFileToRepo(t, repo, "example-single.md", prevTime.Add(time.Duration(prNum)*time.Minute), prNum, "")
	prNum++
	copyFileToRepo(t, repo, "example-multi.md", prevTime.Add(time.Duration(prNum)*time.Minute), prNum, "")
	var last plumbing.Hash
	_, err := repo.CreateTag(cfg.Tag, last, nil)
	requireNoError(t, err)
	merged, err := changelog.Release(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}
	exp, err := os.ReadFile("testdata/expected-examples-merged.md")
	requireNoError(t, err)
	if string(exp) != merged {
		t.Fatalf("expected %s, got %s", exp, merged)
	}
}

func setupTestRepo(t *testing.T) (*git.Repository, *changelog.Config, time.Time, int) {
	prNum := 0
	storage := memory.NewStorage()
	mem := memfs.New()
	repo, err := git.Init(storage, mem)
	requireNoError(t, err)
	prevTime, err := time.Parse("2006-01-02 15:04:05", "2021-01-01 00:00:00")
	requireNoError(t, err)
	copyFileToRepo(t, repo, "previous.md", prevTime, prNum, "v1.0.0")
	releaseTime, err := time.Parse("2006-01-02 15:04:05", "2021-11-11 11:11:11")
	requireNoError(t, err)
	cfg := &changelog.Config{
		ReleaseTime:  releaseTime,
		Repository:   repo,
		ChangesDir:   "changelog",
		Tag:          "v1.0.1",
		PreviousPath: "changelog/previous.md",
		RepoConfig:   changelog.RepoConfig{Owner: "prysmaticlabs", Repo: "prysm"},
	}
	return repo, cfg, prevTime, prNum
}

func TestComplete(t *testing.T) {
	repo, cfg, prevTime, prNum := setupTestRepo(t)
	// add the override fixture to make sure we leave pr links for overrides alone
	prNum++
	copyFileToRepo(t, repo, "override.md", prevTime.Add(time.Minute), prNum, "")

	tree, err := repo.Worktree()
	requireNoError(t, err)
	prNum++
	iter := &fixiter{values: changelog.Sections}
	var last plumbing.Hash
	for f := iter.next(); f != nil; f = iter.next() {
		fh, err := tree.Filesystem.Create(f.filename())
		requireNoError(t, err)
		fh.Write([]byte(f.content()))
		fh.Close()
		commitAddTag(t, repo, f.filename(), prNum, prevTime.Add(time.Duration(1+prNum)*time.Minute), "")
		prNum++
	}
	_, err = repo.CreateTag("v1.0.1", last, nil)
	requireNoError(t, err)
	merged, err := changelog.Release(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	// temporarily uncomment this line and run the test to update the fixture.
	//requireNoError(t, os.WriteFile("testdata/expected-release.md", []byte(merged), 0644))
	exp, err := os.ReadFile("testdata/expected-release.md")
	requireNoError(t, err)
	if string(exp) != merged {
		t.Fatalf("expected %s, got %s", exp, merged)
	}
}

var errEnd = errors.New("end of permutation")

// fixiter returns all possible sets of sections headers.
// section headers are always in the same order, but different sections will
// be missing in each combination.
type fixiter struct {
	missing int
	values  []string
}

func (f *fixiter) next() *fixture {
	// terminal condition, we've got all the bits set
	if f.missing+1 == 1<<len(f.values)-1 {
		return nil
	}
	sections := make([]string, 0, len(f.values))
	// all missing is a special case so we'll just ignore it for simplicity
	f.missing += 1
	for i := range f.values {
		if f.missing>>i&1 == 1 {
			continue
		}
		sections = append(sections, f.values[i])
	}
	return &fixture{sections: sections}
}

type fixture struct {
	name     string
	sections []string
}

func (f *fixture) filename() string {
	return fmt.Sprintf("changelog/%s.md", strings.Join(f.sections, "-"))
}

func (f *fixture) content() string {
	name := f.filename()
	body := ""
	for i, s := range f.sections {
		bullets := bullets(name, 1+(i%3))
		body += "\n### " + s + "\n" + bullets + "\n"
	}
	return body
}

func bullets(base string, n int) string {
	body := ""
	for j := 0; j < n; j++ {
		body += fmt.Sprintf("\n- %s %d", base, j)
	}
	return body
}
