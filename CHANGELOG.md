# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add `global.podSecurityStandards.enforced` value for PSS migration.

## [0.2.0] - 2022-10-13

### Added

- Release app to `gcp-app-collection`

### Changed

- Remove annotation for enabling workload identity. This will apply it to all clusters. We do this because default apps rely on workload-idenity

## [0.1.0] - 2022-10-11

### Added

- Add integration tests
- Add initial reconciler. Copied over from `workload-identity-operator-gcp` 
- Added delete to reconciler

[Unreleased]: https://github.com/giantswarm/fleet-membership-operator-gcp/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/giantswarm/fleet-membership-operator-gcp/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/fleet-membership-operator-gcp/releases/tag/v0.1.0
