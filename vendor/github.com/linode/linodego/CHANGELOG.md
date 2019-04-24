# Change Log

<a name="v0.5.1"></a>
## [v0.5.1](https://github.com/linode/linodego/compare/v0.5.0...v0.5.1) (2018-09-10)

### Fixes

* Domain.Status was not imported from API responses correctly

<a name="v0.5.0"></a>
## [v0.5.0](https://github.com/linode/linodego/compare/v0.4.0...v0.5.0) (2018-09-09)

### Breaking Changes

* List functions return slice of thing instead of slice of pointer to thing

### Feature

* add SSHKeys methods to client (also affects InstanceCreate, InstanceDiskCreate)
* add RebuildNodeBalancerConfig (and CreateNodeBalancerConfig with Nodes)

### Fixes

* Event.TimeRemaining wouldn't parse all possible API value
* Tests no longer rely on known/special instance and volume ids

<a name="0.4.0"></a>
## [0.4.0](https://github.com/linode/linodego/compare/v0.3.0...0.4.0) (2018-08-27)

### Breaking Changes

Replaces bool, error results with error results, for:

* instance\_snapshots.go: EnableInstanceBackups
* instance\_snapshots.go: CancelInstanceBackups
* instance\_snapshots.go: RestoreInstanceBackup
* instances.go: BootInstance
* instances.go: RebootInstance
* instances.go: MutateInstance
* instances.go: RescueInstance
* instances.go: ResizeInstance
* instances.go: ShutdownInstance
* volumes.go: DetachVolume
* volumes.go: ResizeVolume


### Docs

* reword text about breaking changes until first tag

### Feat

* added MigrateInstance and InstanceResizing from 4.0.1-4.0.3 API Changelog
* added gometalinter to travis builds
* added missing function and type comments as reported by linting tools
* supply json values for all fields, useful for mocking responses using linodego types
* use context channels in WaitFor\* functions
* add LinodeTypeClass type (enum)
* add TicketStatus type (enum)
* update template thing and add a test template

### Fix

* TransferQuota was TransferQuote (and not parsed from the api correctly)
* stackscripts udf was not parsed correctly
* add InstanceCreateOptions.PrivateIP
* check the WaitFor timeout before sleeping to avoid extra sleep
* various linting warnings and unhandled err results as reported by linting tools
* fix GetStackscript 404 handling


<a name="0.3.0"></a>

## [0.3.0](https://github.com/linode/linodego/compare/v0.2.0...0.3.0) (2018-08-15)

### Breaking Changes

* WaitForVolumeLinodeID return fetch volume for consistency with out WaitFors
* Moved linodego from chiefy to github.com/linode. Thanks [@chiefy](https://github.com/chiefy)!

<a name="v0.2.0"></a>

## [v0.2.0](https://github.com/linode/linodego/compare/v0.1.1...v0.2.0) (2018-08-11)

### Breaking Changes

* WaitFor\* should be client methods
  *use `client.WaitFor...` rather than `linodego.WaitFor(..., client, ...)`*

* remove ListInstanceSnapshots (does not exist in the API)
  *this never worked, so shouldn't cause a problem*

* Changes UpdateOptions and CreateOptions and similar Options parameters to values instead of pointers
  *these were never optional and the function never updated any values in the Options structures*

* fixed various optional/zero Update and Create options
  *some values are now pointers, and vice-versa*

  * Changes InstanceUpdateOptions to use pointers for optional fields Backups and Alerts
  * Changes InstanceClone's Disks and Configs to ints instead of strings

* using new enum string aliased types where appropriate
  *`InstanceSnapshotStatus`, `DiskFilesystem`, `NodeMode`*

### Feature

* add RescueInstance and RescueInstanceOptions
* add CreateImage, UpdateImage, DeleteImage
* add EnableInstanceBackups, CancelInstanceBackups, RestoreInstanceBackup
* add WatchdogEnabled to InstanceUpdateOptions

### Fix

* return Volume from AttachVolume instead of bool
* add more boilerplate to template.go
* nodebalancers and domain records had no pagination support
* NodeBalancer transfer stats are not int

### Tests

* add fixtures and tests for NodeBalancerNodes
* fix nodebalancer tests to handle changes due to random labels
* add tests for nodebalancers and nodebalancer configs
* added tests for Backups flow
* TestListInstanceBackups fixture is hand tweaked because repeated polled events
  appear to get the tests stuck

### Deps

* update all dependencies to latest

<a name="v0.1.1"></a>

## [v0.1.1](https://github.com/linode/linodego/compare/v0.0.1...v0.1.0) (2018-07-30)

Adds more Domain handling

### Fixed

* go-resty doesnt pass errors when content-type is not set
* Domain, DomainRecords, tests and fixtures

### Added

* add CreateDomainRecord, UpdateDomainRecord, and DeleteDomainRecord

<a name="v0.1.0"></a>

## [v0.1.0](https://github.com/linode/linodego/compare/v0.0.1...v0.1.0) (2018-07-23)

Deals with NewClient and context for all http requests

### Breaking Changes

* changed `NewClient(token, *http.RoundTripper)` to `NewClient(*http.Client)`
* changed all `Client` `Get`, `List`, `Create`, `Update`, `Delete`, and `Wait` calls to take context as the first parameter

### Fixed

* fixed docs should now show Examples for more functions

### Added

* added `Client.SetBaseURL(url string)`

<a name="v0.0.1"></a>
## v0.0.1 (2018-07-20)

### Changed

* Initial tagged release
