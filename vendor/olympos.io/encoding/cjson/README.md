# cjson

[![GoDoc](https://godoc.org/olympos.io/encoding/cjson?status.svg)](https://godoc.org/olympos.io/encoding/cjson)

A Go library to emit JSON in normal/canonical form.

```go
import "olympos.io/encoding/cjson"
```

To use this library, import the import path above with your favourite dependency
management tool and refer to it in the files where you want to use it.

This repository also contains the tool `json_canonicalize`, which canonicalizes
JSON data. You can fetch that one by running

```shell
$ go get olympos.io/encoding/cjson/cmd/json_canonicalize
```

## Usage

The core of this library is a single function:

```go
func Canonicalize(dst io.Writer, src io.Reader) (int, error) {}
```

`Canonicalize` takes a stream of JSON values, canonicalizes them, then sends
them to `dst`.

There are two functions that works on top of `Canonicalize` to ease the work it
does:

```go
func Marshal(v interface{}) ([]byte, error) {}
func NewEncoder(w io.Writer) *Encoder {}
```

Those two works almost as a drop-in replacement for the `encoding/json`
functions with the same name.

## Rationale

JSON can be emitted in many shapes and forms. This is good, but it hinders you
from checking whether two JSON values are identical without reading them into a
structure. If you instead encode them into a canonical form, you can store that
value and efficiently hash it.

## But Why Another One?

There have been previous attempts at creating a formal spec for canonical JSON,
perhaps most notably [Staykov and Hu's draft on JSON Canonical Form](https://tools.ietf.org/html/draft-staykov-hu-json-canonical-form-00).

However, this draft does not consider two things:

1. Canonicalization of strings
2. Integer vs. double

Strings can contain unicode characters that may or may not be written in up to
multiple different ways. Consider, for example, the symbol `/`. In JSON, `/` may
be written as `/`, as `\/`, as `\u002F` and `\u002f`.

The integer vs. double consideration has more with real world usage
consideration: A number in JSON can be considered an integer if it is within the
range `[-(2**53)+1, (2**53)-1]` and has no fractional part. However, many
languages – Go included – will not accept a number that is an integer but
"presented" as a float, for example `1.0E0` (typically the other way around
works fine).

Other JSON specifications exist and attempts to take this into consideration.
Another known speicfication is [Canonical
JSON](http://wiki.laptop.org/go/Canonical_JSON). However, it

1. Disallows floating point values
2. Allows strings to be arbitrary byte sequences

which means that you can not transform arbitrary JSON into this specification,
and, technically speaking, you can not transform canonical JSON into "valid
JSON" (although this is easy to ensure).

I need to encode floats, meaning that
[`canonical/json`](https://godoc.org/github.com/docker/go/canonical/json) was a
no go for me.

## Spec

The encoding rules are described in detail in the [SPEC.md](SPEC.md).


## License

Copyright © 2018-2019 Jean Niklas L'orange

Distributed under the BSD 3-clause license, which is available in the file
LICENSE.
