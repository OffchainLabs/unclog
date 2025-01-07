## unclog the changelog

The purpose of this tool is to enable a more automated changelog management workflow built around
a directory of changelog "fragments", which can be independently added in PRs and merged together
into the previous changelog at release time.

The repo also demonstrates using a github action together with the `unclog check` subcommand 
to automate ensuring that all PRs include a changelog fragment. 

#### What is a changelog fragment?

A changelog fragment is a tiny markdown file consisting only of section headers, matching the set 
of headers defined by the [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format, followed by
bullet points of changelog entries that should be merged into their respect sections into
a combined release changelog. The following sections are supported:
```
### Added
- for new features.
### Changed
- for changes in existing functionality.
### Deprecated
- for soon-to-be removed features.
### Removed
- for now removed features.
### Fixed
- for any bug fixes.
### Security
- in case of vulnerabilities.

### Ignored
- see below
```

Note that in addition to the sections supported by "keep a changelog" we have added a new section called
"Ignored". This section was added because sometimes we have a valid reason for not adding to the changelog,
like a bug fix for an unreleased/incomplete feature, or a PR that reverts a previously merged but unreleased PR.
When the `Ignored` section is used, the bullet point should be a reason that this PR should be excluded
from the changelog.

The fragment files can contain one section or multiple sections, and each section can have 1 or more bullet points.
For example, [here](/unclog/release/testdata/example-single.md) is a fragment file for a typical change,
with just one bullet added to the `Fixed` section:

```
### Fixed

- Example of a single changelog entry.
```

[Here](unclog/release/testdata/example-multi.md) is an example fragment file for a change, where one or more bullets
have been added to multiple sections:

```
### Fixed

- A bug was fixed.
- Another bug was fixed.

### Security

- The bug fixes resolved a security issue.
```

### Why manage the changelog this way

There are a few benefits to using separate files over other options (simple file; deriving from pr titles):
- Standardize changelog entry style; punctuation and automatically include backrefs to pull requests.
- Reduce the possibility for human error. The release subcommand uses the range of commits between the
  previous release tag (as parsed from the previously released changelog in the repo) and the newly tagged
  release.
- Enables github automated checks for the existence of a changelog edit.
- Avoids frustrating conflicts on the changelog itself.
- Enables more control and flexibility compared to trying to parse everything out of PR metadata. We prototyped
  that approach, but it had a few issues like single-commit PRs, PRs that needed to add multiple sections of changelog.

### Best Practices

- Put files in the `changelog` directory.
- Pick a unique name for the changelog file - an obvious choice would be `<github user name>_<branch name>.md`.
- Avoid editing changelogs; only add new files.
- The tool supports adding as many bullet points as you want to any section you want. But don't make up sections,
  unknown sections will be treated as invalid in the github workflow check.

### Installing tool from private repo

Currently this repo is private, so installing the tool for ez use requires a wonky set of steps:
```
$ pushd /tmp && git clone git@github.com:OffchainLabs/unclog.git && pushd unclog && go install . && popd && popd
```

This will install the `unclog` binary on your gopath.

### Using the tool to make a released changelog

Note: before running the tool, make sure you have tagged the release commit. The tool looks at individual PR commits
to find all the information used in building the changelog.

```
$ cd $PRYSM_REPO_DIR
$ git fetch && git checkout origin/develop
$ git checkout -b update-changelog
$ unclog release -repo=$PRYSM_REPO_DIR -tag=$NEW_RELEASE_TAG -cleanup > $PRYSM_REPO_DIR/CHANGELOG.md
$ git commit -m "updating the changelog for $NEW_RELEASE_TAG release"
```

Note the `-cleanup` tag will `git rm` only for the changelog fragments that were found in commits between the
release tag discovered by parsing the previously released changelog file and the tag specified in the `-tag` argument.
So any changelog fragments that are unreleased will be left alone. At this point you can edit the resulting changelog
file to add any desired high level context about the release.

When the release subcommand parses the previously committed changelog, it discards everything in the file that comes before
the release header where the release version is specified. So if you want your notes to persist into the next release, make
sure to place them between the release header and the changelog sections. The release header is the markdown heading that
looks like this: ```## [v5.2.0](https://github.com/prysmaticlabs/prysm/compare/v5.1.2...v5.2.0)```

### Github workflow

The workflow is currently quite heavy because it has to build the tool before using it to check commits. One advatange of
keeping this tool as a separate repo (under the OffchainLabs org) would be setting up a process to pre-build the binary
and copy it to the action. But building the binary each time also works.

Look at the PRs against this repo for examples of the tool allowing and blocking PRs based on the observed changelog fragments.

### Example in prysm

The [changelog-tool](https://github.com/prysmaticlabs/prysm/tree/changelog-tool) branch in prysm has an example commit 
converting many of the unreleased changelog entries to a single fragment file. The commit to transition to this format
should consist of adding the github workflow files, an update to the contributing doc, and a fragment file to account for
all merged PRs.

I did not want to push a semver tag that could throw off anyone's release automation, and attempting to push my test tag
was rejected by our repo's policies. So you need to recreate the tag locally. Here are complete steps for for trying out the
example branch:

```
$ cd $PRYSM_REPO_DIR
$ git fetch && git checkout changelog-tool
$ git tag changelog-test changelog-tool
$ unclog release -repo=$PRYSM_REPO_DIR -tag=$NEW_RELEASE_TAG -cleanup > CHANGELOG.md
$ git status # or git diff, to check out the results
```
