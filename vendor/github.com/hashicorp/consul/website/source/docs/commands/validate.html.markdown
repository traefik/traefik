---
layout: "docs"
page_title: "Commands: Validate"
sidebar_current: "docs-commands-validate"
description: >
  The `consul validate` command tests that config files are valid by
  attempting to parse them. Useful to ensure a configuration change will
  not cause consul to fail after a restart.
---

# Consul Validate

The `consul validate` command performs a basic sanity test on Consul
configuration files. For each file or directory given, the validate command
will attempt to parse the contents just as the "consul agent" command would,
and catch any errors. This is useful to do a test of the configuration only,
without actually starting the agent.

For more information on the format of Consul's configuration files, read the
consul agent [Configuration Files](/docs/agent/options.html#configuration_files)
section.

## Usage

Usage: `consul validate [options] FILE_OR_DIRECTORY...`

Performs a basic sanity test on Consul configuration files. For each file
or directory given, the validate command will attempt to parse the
contents just as the "consul agent" command would, and catch any errors.
This is useful to do a test of the configuration only, without actually
starting the agent.

Returns 0 if the configuration is valid, or 1 if there are problems.

```text
$ consul validate /etc/consul.d
Configuration is valid!
```

