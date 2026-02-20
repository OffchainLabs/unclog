package changelog

import (
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestLogParse(t *testing.T) {
	fixture := `Fix Deadline Again During Rollback (#14686)

	* fix it again

	* CHANGELOG
	`
	var h plumbing.Hash
	o := &object.Commit{
		Hash:    h,
		Message: fixture,
	}
	_, err := parseCommit(o)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseCommitMerge(t *testing.T) {
	fixture := "Merge pull request #42 from org/feature-branch\n\nAdd awesome feature"
	o := &object.Commit{Message: fixture}
	cm, err := parseCommit(o)
	if err != nil {
		t.Fatal(err)
	}
	if cm.pr != 42 {
		t.Fatalf("expected pr 42, got %d", cm.pr)
	}
	if cm.title != "Add awesome feature" {
		t.Fatalf("expected title 'Add awesome feature', got '%s'", cm.title)
	}
}

func TestParseCommitMergeNoBody(t *testing.T) {
	fixture := "Merge pull request #99 from org/hotfix"
	o := &object.Commit{Message: fixture}
	cm, err := parseCommit(o)
	if err != nil {
		t.Fatal(err)
	}
	if cm.pr != 99 {
		t.Fatalf("expected pr 99, got %d", cm.pr)
	}
	if cm.title != "Merge pull request #99 from org/hotfix" {
		t.Fatalf("expected title to be first line, got '%s'", cm.title)
	}
}
