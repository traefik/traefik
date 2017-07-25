# client-go

Go clients for talking to a [kubernetes](http://kubernetes.io/) cluster.

## Table of Contents
 
- [What's included](#whats-included)
- [Versioning](#versioning)
  - [Compatibility: your code <-> client-go](#compatibility-your-code---client-go)
  - [Compatibility: client-go <-> Kubernetes clusters](#compatibility-client-go---kubernetes-clusters)
  - [Compatibility matrix](#compatibility-matrix)
  - [Why do the 1.4 and 1.5 branch contain top-level folder named after the version?](#why-do-the-14-and-15-branch-contain-top-level-folder-named-after-the-version)
- [How to get it](#how-to-get-it)
- [How to use it](#how-to-use-it)
- [Dependency management](#dependency-management)
- [Reporting bugs](#reporting-bugs)
- [Contributing code](#contributing-code)

### What's included

* The `kubernetes` package contains the clientset to access Kubernetes API.
* The `discovery` package is used to discover APIs supported by a Kubernetes API server.
* The `dynamic` package contains a dynamic client that can perform generic operations on arbitrary Kubernetes API objects.
* The `transport` package is used to set up auth and start a connection.
* The `tools/cache` package is useful for writing controllers.

### Versioning

`client-go` version numbers are unrelated to Kubernetes version numbers. Please see the [compatibility matrix](#compatibility-matrix) for the compatible Kubernetes clusters.

`client-go` follows [semver](http://semver.org/). We will not make backwards-incompatible changes without incrementing the major version number. A change is backwards-incompatible either if it *i)* changes the public interfaces of `client-go`, or *ii)* makes `client-go` incompatible with otherwise supported versions of Kubernetes clusters.

We will create a new branch for each increment in the major version number or minor version number. We will create a new tag for each increment in the patch version number. See [semver](http://semver.org/) for definitions of major, minor, and patch.

The master branch will track the head in the main Kubernetes repo and accumulating changes. We will make a new client-go major/minor/patch version when:
* A new major (very rare) or a new minor version is released in the Kubernetes main repository.
* A feature or a bug fix in the master branch is popular and users want it in a stable branch or with a stable tag. 

#### Compatibility: your code <-> client-go

`client-go` follows [semver](http://semver.org/), so until the major version of client-go gets increased, your code will compile and will continue to work with explicitly supported versions of Kubernetes clusters.

#### Compatibility: client-go <-> Kubernetes clusters

`client-go` versions will be backwards compatible with many Kubernetes clusters. As we increment `client-go` versions, we will note which Kubernetes versions we expect them to work with. 

We do not back-port new Kubernetes features into older clients. If you need a new feature, you are expected to upgrade to the new client. You can check out the [CHANGELOG](./CHANGELOG.md) for notable changes.

#### Compatibility matrix

* **client-go/1.4** is compatible with Kubernetes 1.3 through 1.5; it includes all features provided by Kubernetes 1.4.
* **client-go/1.5** is compatible with Kubernetes 1.3 through 1.5; it includes all features provided by Kubernetes 1.4.

#### Why do the 1.4 and 1.5 branch contain top-level folder named after the version?

1.4 and 1.5 branch keep the top-level folders so they do not break the import lines of existing users. These top-level folders are deprecated and are removed from the master branch and future branches.

### How to get it

If you use `go get` to get client-go, **you will get the unstable master branch!** You can `git checkout` a stable branch.

We recommend using a vendor tool, like [godep](https://github.com/tools/godep), [glide](https://github.com/Masterminds/glide), or [govendor](https://github.com/kardianos/govendor) to track a stable version of `client-go`.

### How to use it

If your application runs in a Pod in the cluster, please refer to the in-cluster [example](examples/in-cluster/main.go), otherwise please refer to the out-of-cluster [example](examples/out-of-cluster/main.go).

### Dependency management

If your application depends on a package that client-go depends on, and you let the Go compiler find the dependency in `GOPATH`, you will end up with duplicated dependencies: one copy from the `GOPATH`, and one from the vendor folder of client-go. This will cause unexpected runtime error like flag redefinition, since the go compiler ends up importing both packages separately, even if they are exactly the same thing. If this happens, you can either
* run `godep restore` ([godep](https://github.com/tools/godep)) in the client-go/ folder, then remove the vendor folder of client-go. Then the packages in your GOPATH will be the only copy
* or run `godep save` in your application folder to flatten all dependencies.

### Reporting bugs

Please report bugs to the main Kubernetes [repository](https://github.com/kubernetes/kubernetes/issues/new).

### Contributing code
Please send pull requests against the client packages in the Kubernetes main [repository](https://github.com/kubernetes/kubernetes), and run the `/staging/copy.sh` script to update the staging area in the main repository. Changes in the staging area will be published to this repository every day.
