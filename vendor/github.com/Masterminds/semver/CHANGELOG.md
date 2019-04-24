# Release 1.2.2 (2016-12-13)

## Fixed
- #34: Fixed issue where hyphen range was not working with pre-release parsing.

# Release 1.2.1 (2016-11-28)

## Fixed
- #24: Fixed edge case issue where constraint "> 0" does not handle "0.0.1-alpha"
  properly.

# Release 1.2.0 (2016-11-04)

## Added
- #20: Added MustParse function for versions (thanks @adamreese)
- #15: Added increment methods on versions (thanks @mh-cbon)

## Fixed
- Issue #21: Per the SemVer spec (section 9) a pre-release is unstable and
  might not satisfy the intended compatibility. The change here ignores pre-releases
  on constraint checks (e.g., ~ or ^) when a pre-release is not part of the
  constraint. For example, `^1.2.3` will ignore pre-releases while
  `^1.2.3-alpha` will include them.

# Release 1.1.1 (2016-06-30)

## Changed
- Issue #9: Speed up version comparison performance (thanks @sdboyer)
- Issue #8: Added benchmarks (thanks @sdboyer)
- Updated Go Report Card URL to new location
- Updated Readme to add code snippet formatting (thanks @mh-cbon)
- Updating tagging to v[SemVer] structure for compatibility with other tools.

# Release 1.1.0 (2016-03-11)

- Issue #2: Implemented validation to provide reasons a versions failed a
  constraint.

# Release 1.0.1 (2015-12-31)

- Fixed #1: * constraint failing on valid versions.

# Release 1.0.0 (2015-10-20)

- Initial release
