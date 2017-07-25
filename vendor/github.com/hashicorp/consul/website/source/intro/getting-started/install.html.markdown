---
layout: "intro"
page_title: "Installing Consul"
sidebar_current: "gettingstarted-install"
description: |-
  Consul must first be installed on every node that will be a member of the Consul cluster. To make installation easy, Consul is distributed as a binary package for all supported platforms and architectures. This page will not cover how to compile Consul from source.
---

# Install Consul

Consul must first be installed on your machine. Consul is distributed as a
[binary package](/downloads.html) for all supported platforms and architectures.
This page will not cover how to compile Consul from source, but compiling from
source is covered in the [documentation](/docs/index.html) for those who want to
be sure they're compiling source they trust into the final binary.

## Installing Consul

To install Consul, find the [appropriate package](/downloads.html) for
your system and download it. Consul is packaged as a zip archive.

After downloading Consul, unzip the package. Consul runs as a single binary
named `consul`. Any other files in the package can be safely removed and
Consul will still function.

The final step is to make sure that the `consul` binary is available on the `PATH`.
See [this page](https://stackoverflow.com/questions/14637979/how-to-permanently-set-path-on-linux)
for instructions on setting the PATH on Linux and Mac.
[This page](https://stackoverflow.com/questions/1618280/where-can-i-set-path-to-make-exe-on-windows)
contains instructions for setting the PATH on Windows.

## Verifying the Installation

After installing Consul, verify the installation worked by opening a new
terminal session and checking that `consul` is available. By executing
`consul` you should see help output similar to this:

```text
$ consul
usage: consul [--version] [--help] <command> [<args>]

Available commands are:
    agent          Runs a Consul agent
    event          Fire a new event

# ...
```

If you get an error that `consul` could not be found, your `PATH`
environment variable was not set up properly. Please go back and ensure
that your `PATH` variable contains the directory where Consul was
installed.

## Next Steps

Consul is installed and ready for operation. Let's
[run the agent](agent.html)!
