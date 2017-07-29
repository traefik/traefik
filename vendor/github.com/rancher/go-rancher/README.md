# Go Bindings for Rancher API

# Generating Code
First, you must have a master version of Rancher running. The best way to do this is:
```sh
docker run -p 8080:8080 -d rancher/server:master
```

Once Rancher is running, you can run the gen-schema.sh script:
```sh
./scripts/gen-schema.sh http://<docker host ip>:8080

# The default url is http://localhost:8080, so if rancher/server is listening on localhost, you can omit the url:
./scripts/gen-schema.sh
```

This will add, remove, and modify go files appropriately. Submit a PR that includes *all* these changes.

## Important caveats
1. If you are running on macOS, you must have gnu-sed installed as sed for this to work properly.
2. If you are running against cattle that is running out of an IDE and you don't have go-machine-service running (you probably don't), you'll see a number of unexpected removed or modified files like `generated_host.go` `generated_machine.go` and `generated_*config.go`.

# Building

```sh
godep go build ./client
```

# Tests

```sh
godep go test ./client
```
# Contact
For bugs, questions, comments, corrections, suggestions, etc., open an issue in
 [rancher/rancher](//github.com/rancher/rancher/issues) with a title starting with `[go-rancher] `.

Or just [click here](//github.com/rancher/rancher/issues/new?title=%5Bgo-rancher%5D%20) to create a new issue.


# License
Copyright (c) 2014-2015 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

