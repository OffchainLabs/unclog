### Fixed

- Bullets that wrap onto following lines in a fragment file are now joined into a single changelog entry instead of keeping only the first line. Previously the continuation lines were dropped and a period was appended to the truncated first line (e.g. Prysm v7.2.0 "Both codegen paths (...) stamp the.").
- The release output keeps the previous changelog's trailing newline instead of dropping it, so a release no longer rewrites the last line of CHANGELOG.md.
