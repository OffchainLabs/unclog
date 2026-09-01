### Fixed

- `unclog release` no longer fails on commits without a PR reference (direct pushes, `Merge branch ...` commits). Changelog fragments found in such commits are linked by commit hash.
- Collect every changelog fragment a commit adds instead of only the first, so fragments from PRs that reach the release branch through a branch merge are no longer silently dropped.

### Added

- `unclog release` warns when it drops bullets under a section that is not in the configured section list (e.g. a fragment using `### Fix` instead of `### Fixed`).

### Changed

- Fragments are read as of the commit that introduced them (the PR's merge commit, so edits made during the PR are included). Deleting or editing a fragment after its PR merges (e.g. release cleanup between merge and release) no longer affects the entry. This also applies to reverted PRs: their entries now stay in the generated changelog, with the revert visible only through the reverting PR's own fragment.
