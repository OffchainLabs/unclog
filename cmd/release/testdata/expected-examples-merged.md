# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## [v1.0.1](https://github.com/prysmaticlabs/prysm/compare/v1.0.0...v1.0.1) - 2021-11-11

### Fixed

- Example of a single changelog entry. [[PR]](https://github.com/prysmaticlabs/prysm/pull/1)
- A bug was fixed. [[PR]](https://github.com/prysmaticlabs/prysm/pull/2)
- Another bug was fixed. [[PR]](https://github.com/prysmaticlabs/prysm/pull/2)

### Security

- The bug fixes resolved a security issue. [[PR]](https://github.com/prysmaticlabs/prysm/pull/2)

## [v1.0.0](https://github.com/prysmaticlabs/prysm/compare/v5.1.2...v5.1.3) - 2024-12-04

This is a hotfix release with one change.

Prysm v5.1.1 contains an updated implementation of the beacon api streaming events endpoint. This
new implementation contains a bug that can cause a panic in certain conditions. The issue is
difficult to reproduce reliably and we are still trying to determine the root cause, but in the
meantime we are issuing a patch that recovers from the panic to prevent the node from crashing.

This only impacts the v5.1.1 release beacon api event stream endpoints. This endpoint is used by the
prysm REST mode validator (a feature which requires the validator to be configured to use the beacon
api intead of prysm's stock grpc endpoints) or accessory software that connects to the events api,
like https://github.com/ethpandaops/ethereum-metrics-exporter

### Added

- Add error counter for SSE endpoint. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14681)
- Add error count prom metric. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14670)
- Improve connection/disconnection logging. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14665)
- Better attestation packing for Electra. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14534)
- Add proto for `DataColumnIdentifier`, `DataColumnSidecar`, `DataColumnSidecarsByRangeRequest` and `MetadataV2`. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14649)
- validator REST: attestation v2. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14633)
- feat(issue-12348): add validator index label to validator_statuses me…. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14473)
- validator REST API: block v2 and Electra support. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14623)
- Add `POST /eth/v2/beacon/pool/attestations endpoint`. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14621)
- Add `/eth/v2/validator/aggregate_attestation`. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14481)
- Benchmark process slots. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14616)
- Allow Protobuf State To Be Created Without Copying. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14613)
- Simplify EjectedValidatorIndices. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14588)
- Rollback Block During Processing. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14554)
- Add `GET /eth/v2/beacon/pool/attestations` endpoint. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14560)
- Use engine api `get-blobs` for block subscriber. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14513)
- Update the monitor package to Electra. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14562)
- Add `/eth/v2/validator/aggregate_and_proofs`. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14490)
- Execution API Electra: requests as a sidecar. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14492)

### Changed

- Add metadata fields to getBlobSidecars. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14677)
- http response handling improvements. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14673)
- Use slot to determine fork version. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14653)
- Check if validator exists when applying pending deposit. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14666)
- Update light-client consensus types. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14652)
- Add missing Eth-Consensus-Version headers. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14647)
- Update light client protobufs. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14650)
- defer payload attribute computation. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14644)
- reorganizing p2p and backfill service registration for consistency. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14640)
- Rename remaining "deposit receipt" to "deposit request". [[PR]](https://github.com/prysmaticlabs/prysm/pull/14629)
- Optimize Message ID Computation. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14591)
- Return early blob constructor if not deneb. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14605)
- Fix various small things in state-native code. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14604)
- prevent panic by returning on connection error. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14602)
- Blocks after capella are execution. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14614)
- Build Proto State Object Once. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14612)
- Change the signature of ProcessPayload. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14610)
- Use ROBlock earlier in the pipeline. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14609)
- Fix order & add slashed=false to new validator instances. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14595)
- Electra: exclude empty requests in requests list. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14580)
- Simplify ExitedValidatorIndices. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14587)
- Use Read Only Head State When Computing Active Indices. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14592)
- Update beacon-chain pgo profile. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14589)
- Add missing version headers. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14566)
- Use ROBlock in block processing pipeline. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14571)
- Update CHANGELOG.md to reflect v5.1.2 hotfix release. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14547)
- Use read only validator for processing. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14558)
- Rollback on errors from forkchoice insertion. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14556)
- rollback on SaveState error. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14555)
- Update correlation penalty for EIP-7251. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14456)

### Removed

- Remove outdated spectest exclusions for EIP-6110. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14630)
- Remove validator count log. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14600)

### Fixed

- Fix Deadline Again During Rollback. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14686)
- `listenForNewNodes` and `FindPeersWithSubnet`: Stop using `ReadNodes` and use iterator instead. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14669)
- chore: fix 404 status URL. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14675)
- Remove kzg proof check for blob reconstructor. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14671)
- Diverse log improvements, comment additions and small refactors. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14658)
- Fix eventstream electra atts. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14655)
- adding nil checks on attestation interface. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14638)
- move get data after nil check for attestations. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14642)
- Validator REST api: adding in check for empty keys changed. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14637)
- Electra: unskipping merkle spec tests:. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14635)
- Use GetBlockAttestationV2 at handler. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14624)
- Rollback Block With Context Deadline. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14608)
- adding in a check to make sure duplicates are now allowed. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14601)
- eip7251: Bugfix and more withdrawal tests. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14578)
- small improvements to logs. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14405)
- keymanager API: bug fixes and inconsistencies. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14586)
- fix --backfill-oldest-slot flag handling. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14584)
- Fix length check between kzg commitments and exist. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14581)
- Docker fix: Update bazel-lib to latest version. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14579)
- Safe StreamEvents write loop. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14557)
- recover from panics when writing the event stream. [[PR]](https://github.com/prysmaticlabs/prysm/pull/14545)