# Contributing to go-ovh

## Submitting Modifications:

So you want to contribute you work? Awesome! We are eager to review it.
To submit your contribution, you must use Github Pull Requests. Your work
does not need to be fully polished before submiting it. Actully, we love
helping people writing a great contribution. Hence, if you are wondering
how to integrate a specific change, feel free to start a discussion in
a Pull Request.

Before we can actually accept and merge a Pull Request, it will need
to follow the conding guidelines (see below), and each commit shall be
signed to indicate your full agreement with these guidelines and the
DCO (see below).

To sign a commit, you may use a command like:

```
# New commit
git commit -s

# Previous commit
git commit --amend -s
```

If a Pull Request can not be automatically merged, you will probably need
to "rebase" your work on latest project update:

```
# Assuming, this project remote is registered as "upstream"
git fetch upstream
git rebase upstream/master
```

## Contribution guidelines

1. your code must follow the coding style rules (see below)
2. your code must be documented
3. you code must be tested
4. your work must be signed (see "Developer Certificate of Origin" below)
5. you may contribute through GitHub Pull Requests

## Coding and documentation Style:

- Code must be formated with `gofmt -sw ./`
- Code must pass `go vet ./...`
- Code must pass `golint ./...`

## Licensing for new files:

go-ovh is licensed under a (modified) BSD license. Anything contributed to
go-ovh must be released under this license.

When introducing a new file into the project, please make sure it has a
copyright header making clear under which license it''s being released.

## Developer Certificate of Origin:

```
To improve tracking of contributions to this project we will use a
process modeled on the modified DCO 1.1 and use a "sign-off" procedure
on patches that are being contributed.

The sign-off is a simple line at the end of the explanation for the
patch, which certifies that you wrote it or otherwise have the right
to pass it on as an open-source patch.  The rules are pretty simple:
if you can certify the below:

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I have
    the right to submit it under the open source license indicated in
    the file; or

(b) The contribution is based upon previous work that, to the best of
    my knowledge, is covered under an appropriate open source License
    and I have the right under that license to submit that work with
    modifications, whether created in whole or in part by me, under
    the same open source license (unless I am permitted to submit
    under a different license), as indicated in the file; or

(c) The contribution was provided directly to me by some other person
    who certified (a), (b) or (c) and I have not modified it.

(d) The contribution is made free of any other party''s intellectual
    property claims or rights.

(e) I understand and agree that this project and the contribution are
    public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.


then you just add a line saying

    Signed-off-by: Random J Developer <random@developer.org>

using your real name (sorry, no pseudonyms or anonymous contributions.)
```

