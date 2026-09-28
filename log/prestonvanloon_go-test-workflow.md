### Changed

- Added a `test` GitHub Actions workflow that runs `go vet` and `go test -race` on pull requests and pushes to `main`; previously no workflow ran the test suite.
