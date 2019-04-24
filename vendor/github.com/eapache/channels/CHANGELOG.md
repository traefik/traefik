# Changelog

#### Version 1.1.0 (2015-11-22)

Bug Fixes:
 - The `Len()` and `Cap()` methods on several implementations were racy
   ([#18](https://github.com/eapache/channels/issues/18)).

Note: Fixing the above issue led to a fairly substantial performance hit
(anywhere from 10-25% in benchmarks depending on use case) and involved fairly
major refactoring, which is why this is being released as v1.1.0 instead
of v1.0.1.

#### Version 1.0.0 (2015-01-24)

Version 1.0.0 is the first tagged release. All core functionality was available
at this point.
