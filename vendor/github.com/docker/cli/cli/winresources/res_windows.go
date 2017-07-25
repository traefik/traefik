/*Package winresources is used to embed Windows resources into docker.exe.
These resources are used to provide

    * Version information
    * An icon
    * A Windows manifest declaring Windows version support

The resource object files are generated with go generate.
The resource source files are located in scripts/winresources.
This occurs automatically when you run scripts/build/windows.

These object files are picked up automatically by go build when this package
is included.

*/
package winresources

//go:generate ../../scripts/gen/windows-resources
