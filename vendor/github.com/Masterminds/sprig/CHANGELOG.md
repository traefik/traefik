# Changelog

## Release 2.15.0 (2018-04-02)

### Added

- #68 and #69: Add json helpers to docs (thanks @arunvelsriram)
- #66: Add ternary function (thanks @binoculars)
- #67: Allow keys function to take multiple dicts (thanks @binoculars)
- #89: Added sha1sum to crypto function (thanks @benkeil)
- #81: Allow customizing Root CA that used by genSignedCert (thanks @chenzhiwei)
- #92: Add travis testing for go 1.10
- #93: Adding appveyor config for windows testing

### Changed

- #90: Updating to more recent dependencies
- #73: replace satori/go.uuid with google/uuid (thanks @petterw)

### Fixed

- #76: Fixed documentation typos (thanks @Thiht)
- Fixed rounding issue on the `ago` function. Note, the removes support for Go 1.8 and older

## Release 2.14.1 (2017-12-01)

### Fixed

- #60: Fix typo in function name documentation (thanks @neil-ca-moore)
- #61: Removing line with {{ due to blocking github pages genertion
- #64: Update the list functions to handle int, string, and other slices for compatibility

## Release 2.14.0 (2017-10-06)

This new version of Sprig adds a set of functions for generating and working with SSL certificates.

- `genCA` generates an SSL Certificate Authority
- `genSelfSignedCert` generates an SSL self-signed certificate
- `genSignedCert` generates an SSL certificate and key based on a given CA

## Release 2.13.0 (2017-09-18)

This release adds new functions, including:

- `regexMatch`, `regexFindAll`, `regexFind`, `regexReplaceAll`, `regexReplaceAllLiteral`, and `regexSplit` to work with regular expressions
- `floor`, `ceil`, and `round` math functions
- `toDate` converts a string to a date
- `nindent` is just like `indent` but also prepends a new line
- `ago` returns the time from `time.Now`

### Added

- #40: Added basic regex functionality (thanks @alanquillin)
- #41: Added ceil floor and round functions (thanks @alanquillin)
- #48: Added toDate function (thanks @andreynering)
- #50: Added nindent function (thanks @binoculars)
- #46: Added ago function (thanks @slayer)

### Changed

- #51: Updated godocs to include new string functions (thanks @curtisallen)
- #49: Added ability to merge multiple dicts (thanks @binoculars)

## Release 2.12.0 (2017-05-17)

- `snakecase`, `camelcase`, and `shuffle` are three new string functions
- `fail` allows you to bail out of a template render when conditions are not met

## Release 2.11.0 (2017-05-02)

- Added `toJson` and `toPrettyJson`
- Added `merge`
- Refactored documentation

## Release 2.10.0 (2017-03-15)

- Added `semver` and `semverCompare` for Semantic Versions
- `list` replaces `tuple`
- Fixed issue with `join`
- Added `first`, `last`, `intial`, `rest`, `prepend`, `append`, `toString`, `toStrings`, `sortAlpha`, `reverse`, `coalesce`, `pluck`, `pick`, `compact`, `keys`, `omit`, `uniq`, `has`, `without`

## Release 2.9.0 (2017-02-23)

- Added `splitList` to split a list
- Added crypto functions of `genPrivateKey` and `derivePassword`

## Release 2.8.0 (2016-12-21)

- Added access to several path functions (`base`, `dir`, `clean`, `ext`, and `abs`)
- Added functions for _mutating_ dictionaries (`set`, `unset`, `hasKey`)

## Release 2.7.0 (2016-12-01)

- Added `sha256sum` to generate a hash of an input
- Added functions to convert a numeric or string to `int`, `int64`, `float64`

## Release 2.6.0 (2016-10-03)

- Added a `uuidv4` template function for generating UUIDs inside of a template.

## Release 2.5.0 (2016-08-19)

- New `trimSuffix`, `trimPrefix`, `hasSuffix`, and `hasPrefix` functions
- New aliases have been added for a few functions that didn't follow the naming conventions (`trimAll` and `abbrevBoth`)
- `trimall` and `abbrevboth` (notice the case) are deprecated and will be removed in 3.0.0

## Release 2.4.0 (2016-08-16)

- Adds two functions: `until` and `untilStep`

## Release 2.3.0 (2016-06-21)

- cat: Concatenate strings with whitespace separators.
- replace: Replace parts of a string: `replace " " "-" "Me First"` renders "Me-First"
- plural: Format plurals: `len "foo" | plural "one foo" "many foos"` renders "many foos"
- indent: Indent blocks of text in a way that is sensitive to "\n" characters.

## Release 2.2.0 (2016-04-21)

- Added a `genPrivateKey` function (Thanks @bacongobbler)

## Release 2.1.0 (2016-03-30)

- `default` now prints the default value when it does not receive a value down the pipeline. It is much safer now to do `{{.Foo | default "bar"}}`.
- Added accessors for "hermetic" functions. These return only functions that, when given the same input, produce the same output.

## Release 2.0.0 (2016-03-29)

Because we switched from `int` to `int64` as the return value for all integer math functions, the library's major version number has been incremented.

- `min` complements `max` (formerly `biggest`)
- `empty` indicates that a value is the empty value for its type
- `tuple` creates a tuple inside of a template: `{{$t := tuple "a", "b" "c"}}`
- `dict` creates a dictionary inside of a template `{{$d := dict "key1" "val1" "key2" "val2"}}` 
- Date formatters have been added for HTML dates (as used in `date` input fields)
- Integer math functions can convert from a number of types, including `string` (via `strconv.ParseInt`).

## Release 1.2.0 (2016-02-01)

- Added quote and squote
- Added b32enc and b32dec
- add now takes varargs
- biggest now takes varargs

## Release 1.1.0 (2015-12-29)

- Added #4: Added contains function. strings.Contains, but with the arguments
  switched to simplify common pipelines. (thanks krancour)
- Added Travis-CI testing support

## Release 1.0.0 (2015-12-23)

- Initial release
