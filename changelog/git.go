package changelog

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

var prCommitRE = regexp.MustCompile(`^(.*) \(#(\d+)\)$`)

type Commit struct {
	pr    int
	title string
	gc    *object.Commit
}

func (c Commit) Id() string {
	return c.gc.Hash.String()
}

func (c Commit) Parent() (Commit, error) {
	p, err := c.gc.Parent(0)
	if err != nil {
		return Commit{}, err
	}
	return Commit{gc: p}, nil
}

func (c Commit) prLink(repo *RepoConfig) string {
	return "[[PR]](" + repo.PrURL(c.pr) + ")"
}

func tagTimestamp(r *git.Repository, tag string) (time.Time, error) {
	t, err := r.Tag(tag)
	if err != nil {
		return time.Time{}, err
	}
	obj, err := r.CommitObject(t.Hash())
	if err != nil {
		return time.Time{}, err
	}
	return obj.Author.When, nil
}

func getFile(cfg *Config, path string) (io.Reader, io.Closer, error) {
	repo, err := cfg.Repo()
	if err != nil {
		return nil, nil, err
	}
	tree, err := repo.Worktree()
	if err != nil {
		return nil, nil, err
	}
	fh, err := tree.Filesystem.Open(path)
	return fh, fh, err
}

func commitsAfter(cfg *Config) ([]Commit, error) {
	r, err := cfg.Repo()
	if err != nil {
		return nil, err
	}
	since, err := tagTimestamp(r, cfg.Previous.Version)
	if err != nil {
		return nil, err
	}
	// We want to filter out the previous release commit, and the Since filter is inclusive,
	// so we increment it by the smallest amount.
	since = since.Add(time.Nanosecond)
	from, err := r.Tag(cfg.Tag)
	if err != nil {
		return nil, err
	}
	iter, err := r.Log(&git.LogOptions{Since: &since, From: from.Hash()})
	if err != nil {
		return nil, err
	}
	commits := make([]Commit, 0)
	err = iter.ForEach(func(c *object.Commit) error {
		cm, err := parseCommit(c)
		if err != nil {
			return err
		}
		commits = append(commits, cm)
		return nil
	})
	// reverse the list
	for i, j := 0, len(commits)-1; i < j; i, j = i+1, j-1 {
		commits[i], commits[j] = commits[j], commits[i]
	}
	return commits, err
}

func parseCommit(c *object.Commit) (Commit, error) {
	_, firstLine, err := bufio.ScanLines([]byte(c.Message), true)
	if err != nil {
		return Commit{}, err
	}
	first := string(firstLine)
	m := prCommitRE.FindStringSubmatch(first)
	if m == nil {
		return Commit{}, fmt.Errorf("could not parse format of commit message: %s", c.Message)
	}
	pri, err := strconv.Atoi(m[2])
	return Commit{
		title: m[1],
		pr:    pri,
		gc:    c,
	}, err
}

var errNoChangelogFragment = errors.New("no changelog fragment found")

// Fragment represents a changelog fragment file. That it's a markdown file that has a subset of
// the expected list of changelog headers with bullet points for each section.
type Fragment struct {
	Lines  []string
	Path   string
	Commit Commit
}

func cleanupFragments(cfg *Config, fragments []Fragment) error {
	repo, err := cfg.Repo()
	if err != nil {
		return err
	}
	tree, err := repo.Worktree()
	if err != nil {
		return err
	}
	for _, f := range fragments {
		if err := tree.Filesystem.Remove(path.Join(cfg.ChangesDir, f.Path)); err != nil {
			// A changelog fragment might have been removed already and that's OK. This happens when a PR is reverted.
			log.Printf("Skipping fragment: could not remove changelog fragment %s: %v", f.Path, err)
		}
	}
	return nil
}

func FindFragment(clDir string, parent, cm Commit) (Fragment, error) {
	frag := Fragment{Commit: cm}
	pt, err := parent.gc.Tree()
	if err != nil {
		return frag, err
	}
	t, err := cm.gc.Tree()
	if err != nil {
		return frag, err
	}
	changes, err := object.DiffTreeWithOptions(context.Background(), pt, t, object.DefaultDiffTreeOptions)
	if err != nil {
		return frag, err
	}
	for _, ch := range changes {
		from, to, err := ch.Files()
		if err != nil {
			return frag, err
		}
		// For insertions From is the zero value, and these are the only changes we care about.
		if from != nil {
			continue
		}
		if strings.HasPrefix(ch.To.Name, clDir) {
			if path.Ext(ch.To.Name) != ".md" {
				continue
			}
			frag.Lines, err = to.Lines()
			frag.Path = to.Name
			return frag, err
		}
	}
	return frag, errNoChangelogFragment
}

const maxBranchDepth = 60

var errNoCommitsInBranch = errors.New("branch tip commit exists in upstream")
var errBranchParentNotFound = errors.New("could not find commit where branch diverges from upstream")
var errBranchTooDeep = errors.New("unable to find parent of branch within maximum branch depth")

type historyNode struct {
	commit *Commit
	child  *historyNode
}

func (n *historyNode) chain() []*Commit {
	if n == nil {
		return []*Commit{}
	}
	d := []*Commit{n.commit}
	return append(d, n.child.chain()...)
}

type history map[plumbing.Hash]*historyNode

func (h history) addParent(c *Commit) (*Commit, error) {
	parent, err := c.gc.Parent(0)
	if err != nil {
		return nil, err
	}
	p := h.add(parent)
	h.setChild(p, c)
	return p, nil
}

func (h history) setChild(parent *Commit, child *Commit) {
	h[parent.gc.Hash].child = h.get(child.gc.Hash)
}

func (h history) add(c *object.Commit) *Commit {
	h[c.Hash] = &historyNode{commit: &Commit{gc: c}}
	return h[c.Hash].commit
}

func (h history) get(c plumbing.Hash) *historyNode {
	return h[c]
}

func (h history) descendents(c plumbing.Hash) []*Commit {
	parent := h.get(c)
	if parent == nil || parent.child == nil {
		return []*Commit{}
	}
	return parent.child.chain()
}

func (h history) branchDiff(upstream history, c *Commit) (*Commit, []*Commit, error) {
	// check if the child of c also exists in the upstream history
	// if yes, keep searching
	// if not, return the commit and the chain of commits from the child to c
	lastUpstream := c
	for {
		if c == nil {
			return nil, nil, errBranchParentNotFound
		}
		// if the commit is in the upstream, we need to keep searching for the first child that's only on the branch
		if found := upstream.get(c.gc.Hash); found != nil {
			// found is the history node from upstream - we need to grab the branch node to traverse
			bf := h.get(c.gc.Hash)
			if bf.child == nil {
				return c, []*Commit{}, nil
			}
			if bf == nil || bf.child == nil {
				return nil, nil, errBranchParentNotFound
			}
			lastUpstream = c
			c = bf.child.commit
			continue
		}
		// we got here because the commit is *not* in the upstream history -- this is the commit we are looking for.
		return lastUpstream, h.descendents(lastUpstream.gc.Hash), nil
	}
}

// BranchCommits returns a list of commits on the branch relative to the upstream main branch.
// mainRev is the revision of the main branch. For instance the main prysm repo uses "origin/develop".
// branchRev is either HEAD if the branch is checked out, or the branch name.
// note that the return commits don't have a pr or title because these come from the final squash merged commit.
func BranchCommits(cfg *Config, mainRev, branchRev string) (*Commit, []*Commit, error) {
	r, err := cfg.Repo()
	if err != nil {
		return nil, nil, err
	}
	devr, err := r.ResolveRevision(plumbing.Revision(mainRev))
	if err != nil {
		return nil, nil, fmt.Errorf("could not resolve revision %s: %w", mainRev, err)
	}
	upstream, err := r.CommitObject(*devr)
	if err != nil {
		return nil, nil, fmt.Errorf("could not find upstream commit object for hash %x: %w", *devr, err)
	}

	branchr, err := r.ResolveRevision(plumbing.Revision(branchRev))
	if err != nil {
		return nil, nil, fmt.Errorf("could not resolve revision %s: %w", branchRev, err)
	}
	branch, err := r.CommitObject(*branchr)
	if err != nil {
		return nil, nil, fmt.Errorf("could not find branch commit object for hash %x: %w", *branchr, err)
	}

	// compare tips
	if branch.Hash == upstream.Hash {
		return nil, nil, errNoCommitsInBranch
	}

	upHistory := make(history)
	nextUpstream := upHistory.add(upstream)
	branchHistory := make(history)
	nextBranch := branchHistory.add(branch)

	depth := 0
	for {
		if depth > maxBranchDepth {
			return nil, nil, errBranchTooDeep
		}

		// happy path: branch commit is close to the tip of upstream
		// so look at the next parent in the branch and check if it's in the upstream we've seen so far
		nextBranch, err = branchHistory.addParent(nextBranch)
		if err != nil {
			return nil, nil, err
		}
		if found := branchHistory.get(nextUpstream.gc.Hash); found != nil {
			return branchHistory.branchDiff(upHistory, nextUpstream)
		}

		// we could also have a branch that is based on a commit deep in the upstream. in that scenario, we'll see
		// the commit in the branch before in the upstream. so we advance the history on the upstream side and check
		// if that commit has already been seen on the branch.
		nextUpstream, err = upHistory.addParent(nextUpstream)
		if err != nil {
			return nil, nil, err
		}
		if found := upHistory.get(nextBranch.gc.Hash); found != nil {
			return branchHistory.branchDiff(upHistory, nextBranch)
		}
		depth++
	}
}
