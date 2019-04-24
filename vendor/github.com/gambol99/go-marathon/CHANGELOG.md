# Change Log
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/)
and this project adheres to [Semantic Versioning](http://semver.org/).

## [Unreleased]
### Added
- [#273][PR273] Implement readiness checks.
- [#267][PR267] Add DCOS path parameter for additional marathon instances.

## [0.7.1] - 2017-02-20
### Fixed
- [#261][PR261] Fix URL parsing for Go 1.8.

## [0.7.0] - 2017-02-17
### Added
- [#256][PR256] Expose task state.

### Changed
- [#259][PR259] Add 'omitempty' to UpgradeStrategy properties.

## [0.6.0] - 2016-12-14
### Added
- [#246][PR246] Add TaskKillGracePeriodSeconds support.
- [#244][PR244] Add taskStats support.

### Changed
- [#242][PR242] Pointerize IPAddressPerTask.Discovery.

## [0.5.1] - 2016-11-09
### Fixed
- [#239][PR239] Fix scheme-less endpoint with port.

## [0.5.0] - 2016-11-07
### Fixed
- [#231][PR231] Fixed Marathon cluster code
- [#229][PR229] Add LastFailureCause field to HealthCheckResult.

## [0.4.0] - 2016-10-28
### Added
- [#223][PR223] Add support for IP-per-task.
- [#220][PR220] Add external volume definition to container.
- [#211][PR211] Close event channel on event listener removal.

### Fixed
- [#218][PR218] Remove TimeWaitPolling from marathonClient.
- [#214][PR214] Remove extra pointer layers when passing to r.api*.

## [0.3.0] - 2016-09-28
- [#201][PR201]: Subscribe method is now exposed on the client to allow subscription of callback URL's

### Fixed
- [#205][PR205]: Fix memory leak by signalling goroutine termination on event listener removal.

### Changed
- [#205][PR205]: Change AddEventsListener to return event channel instead of taking one.

## [0.2.0] - 2016-09-23
### Added
- [#196][PR196]: Port definitions.
- [#191][PR191]: name and labels to portMappings.

### Changed
- [#191][PR191] ExposePort() now takes a portMapping instance.

### Fixed
- [#202][PR202]: Timeout error in WaitOnApplication.

## [0.1.1] - 2016-09-07
### Fixed
- Drop question mark-only query parameter in Applications(url.Values) manually
  due to changed behavior in Go 1.7's net/url.Parse.

## [0.1.0] - 2016-08-01
### Added
- Field `message` to the EventStatusUpdate struct.
- Method `Host()` to set host mode explicitly.
- Field `port` to HealthCheck.
- Support for launch queues.
- Convenience method `AddFetchURIs()`.
- Support for forced operations across all methods.
- Filtering method variants (`*By`-suffixed).
- Support for Marathon DCOS token.
- Basic auth and HTTP client settings.
- Marshalling of `Deployment.DeploymentStep` for Marathon v1.X.
- Field `ipAddresses` to tasks and events.
- Field `slaveId` to tasks.
- Convenience methods to populate/clear pointerized values.
- Method `ApplicationByVersion()` to retrieve version-specific apps.
- Support for fetch URIs.
- Parse API error responses on all error types for programmatic evaluation.

### Changed
- Consider app as unhealthy in ApplicationOK if health check is missing. (Ensures result stability during all phases of deployment.)
- Various identifiers violating golint rules.
- Do not set "bridged" mode on Docker containers by default.

### Fixed
- Flawed unmarshalling of `CurrentStep` in events.
- Missing omitempty tag modifiers on `Application.Uris`.
- Missing leading slash in path used by `Ping()`.
- Flawed `KillTask()` in case of hierarchical app ID path.
- Missing omitempty tag modifier on `PortMapping.Protocol`.
- Nil dereference on empty debug log.
- Various occasions where omitted and empty fields could not be distinguished.

## 0.0.1 - 2016-01-27
### Added
- Initial SemVer release.

[Unreleased]: https://github.com/gambol99/go-marathon/compare/v0.7.1...HEAD
[0.7.1]: https://github.com/gambol99/go-marathon/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/gambol99/go-marathon/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/gambol99/go-marathon/compare/v0.5.1...v0.6.0
[0.5.1]: https://github.com/gambol99/go-marathon/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/gambol99/go-marathon/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/gambol99/go-marathon/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/gambol99/go-marathon/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/gambol99/go-marathon/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/gambol99/go-marathon/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/gambol99/go-marathon/compare/v0.0.1...v0.1.0

[PR273]: https://github.com/gambol99/go-marathon/pull/273
[PR267]: https://github.com/gambol99/go-marathon/pull/267
[PR261]: https://github.com/gambol99/go-marathon/pull/261
[PR259]: https://github.com/gambol99/go-marathon/pull/259
[PR256]: https://github.com/gambol99/go-marathon/pull/256
[PR246]: https://github.com/gambol99/go-marathon/pull/246
[PR244]: https://github.com/gambol99/go-marathon/pull/244
[PR242]: https://github.com/gambol99/go-marathon/pull/242
[PR239]: https://github.com/gambol99/go-marathon/pull/239
[PR231]: https://github.com/gambol99/go-marathon/pull/231
[PR229]: https://github.com/gambol99/go-marathon/pull/229
[PR223]: https://github.com/gambol99/go-marathon/pull/223
[PR220]: https://github.com/gambol99/go-marathon/pull/220
[PR218]: https://github.com/gambol99/go-marathon/pull/218
[PR214]: https://github.com/gambol99/go-marathon/pull/214
[PR211]: https://github.com/gambol99/go-marathon/pull/211
[PR205]: https://github.com/gambol99/go-marathon/pull/205
[PR202]: https://github.com/gambol99/go-marathon/pull/202
[PR201]: https://github.com/gambol99/go-marathon/pull/201
[PR196]: https://github.com/gambol99/go-marathon/pull/196
[PR191]: https://github.com/gambol99/go-marathon/pull/191
