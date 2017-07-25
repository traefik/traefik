# zbase32 -- Human-oriented encoding for binary data

Package `zbase32` implements the `z-base-32` encoding as specified in
http://philzimmermann.com/docs/human-oriented-base-32-encoding.txt

This package has been extensively tested to match the behavior of the
[zbase32 Python package](https://pypi.python.org/pypi/zbase32/).

Note that this is **not** RFC 4648/3548, for that see
[encoding/base32](http://golang.org/pkg/encoding/base32/). `z-base-32`
is a variant that aims to be more human-friendly, and in some
circumstances shorter.

For usage, see [godoc](http://godoc.org/pkg/github.com/tv42/zbase32/).

## Command line utilities

Included are simple command-line utilities for encoding/decoding data.
Example:

```console
$ echo hello, world | zbase32-encode
pb1sa5dxfoo8q551pt1yw
$ zbase32-decode pb1sa5dxfoo8q551pt1yw
hello, world
$ printf '\x01binary!!!1\x00' | zbase32-encode
yftg15ubqjh1nejbgryy
$ zbase32-decode yftg15ubqjh1nejbgryy | hexdump -C
00000000  01 62 69 6e 61 72 79 21  21 21 31 00              |.binary!!!1.|
0000000c
```
