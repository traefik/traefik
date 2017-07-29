# Change Log

**ATTN**: This project uses [semantic versioning](http://semver.org/).

## [Unreleased]
### Added
- `Recovery.ErrorHandlerFunc` for custom error handling during recovery
- `With()` helper for building a new `Negroni` struct chaining handlers from
  existing `Negroni` structs

### Fixed
- `Written()` correct returns `false` if no response header has been written

### Changed
- Set default status to `0` in the case that no handler writes status -- was
  previously `200` (in 0.2.0, before that it was `0` so this reestablishes that
  behavior)
- Catch `panic`s thrown by callbacks provided to the `Recovery` handler

## [0.2.0] - 2016-05-10
### Added
- Support for variadic handlers in `New()`
- Added `Negroni.Handlers()` to fetch all of the handlers for a given chain
- Allowed size in `Recovery` handler was bumped to 8k
- `Negroni.UseFunc` to push another handler onto the chain

### Changed
- Set the status before calling `beforeFuncs` so the information is available to them
- Set default status to `200` in the case that no handler writes status -- was previously `0`
- Panic if `nil` handler is given to `negroni.Use`

## 0.1.0 - 2013-07-22
### Added
- Initial implementation.

[Unreleased]: https://github.com/urfave/negroni/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/urfave/negroni/compare/v0.1.0...v0.2.0
