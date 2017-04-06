/*
Package ty provides utilities for writing type parametric functions with run
time type safety.

This package contains two sub-packages `fun` and `data` which define some
potentially useful functions and abstractions using the type checker in
this package.

Requirements

Go tip (or 1.1 when it's released) is required. This package will not work
with Go 1.0.x or earlier.

The very foundation of this package only recently became possible with the
addition of 3 new functions in the standard library `reflect` package:
SliceOf, MapOf and ChanOf. In particular, it provides the ability to
dynamically construct types at run time from component types.

Further extensions to this package can be made if similar functions are added
for structs and functions(?).
*/
package ty
