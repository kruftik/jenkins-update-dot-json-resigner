// Copyright 2018-2019 Jean Niklas L'orange. All rights reserved.
// Use of this source code is governed by a BSD-3-Clause
// license that can be found in the LICENSE file.

// Package cjson canonicalizes JSON. This package implements the
// canonicalization rules defined in
// https://github.com/olympos-labs/cjson/blob/master/SPEC.md and provides
// utility functions around converting both existing and new data to canonical
// JSON.
//
// In addition, the package contains a small command line tool named
// json_canonicalize. It can be installed by calling
//     go get olympos.io/encoding/cjson/cmd/json_canonicalize
package cjson // import "olympos.io/encoding/cjson"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
)

// Marshal returns the canonical JSON encoding of v.
//
// See the documentation for encoding/json.Marshal for details about the
// conversion of Go values to JSON.
func Marshal(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	err := NewEncoder(&buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// An Encoder writes canonical JSON values to an output stream.
type Encoder struct {
	c        canonicalizer
	dst      io.Writer
	setSpace bool
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		dst:      w,
		setSpace: true,
	}
}

// Encode writes the JSON encoding of v to the stream. If multiple values are
// encoded, they will be separated by a space if necessary if StreamSpace is
// enabled (default on).
//
// See the documentation for encoding/json.Marshal for details about the
// conversion of Go values to JSON.
func (e *Encoder) Encode(v interface{}) error {
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	// Wait what
	e.c.dec = json.NewDecoder(bytes.NewBuffer(bs))
	_, err = e.c.value(e.dst)
	e.c.needsSpace = e.c.needsSpace && e.setSpace
	return err
}

// SetStreamSpace turns spaces between values that must be separated on or off.
//
// You should typically not modify this unless you manually manage stream
// separation.
func (e *Encoder) SetStreamSpace(on bool) {
	e.setSpace = on
}

// Canonicalize canonicalizes JSON values from src and puts the result into dst.
// Multiple values from src will be separated by a space only if necessary.
func Canonicalize(dst io.Writer, src io.Reader) (int64, error) {
	c := &canonicalizer{
		dec: json.NewDecoder(src),
	}
	var written int64
	var err error
	for {
		var w int64
		w, err = c.value(dst)
		if err != nil {
			break
		}
		written += w
	}
	if err == io.EOF {
		err = nil
	}
	return written, err
}

type canonicalizer struct {
	scratch    bytes.Buffer
	dec        *json.Decoder
	needsSpace bool
}

func (c *canonicalizer) value(dst io.Writer) (int64, error) {
	tok, err := c.dec.Token()
	if err != nil {
		return 0, err
	}
	switch tok := tok.(type) {
	case string:
		c.needsSpace = false
		return c.writeString(dst, tok)
	case float64:
		var written int64
		if c.needsSpace {
			w, err := dst.Write([]byte{' '})
			written += int64(w)
			if err != nil {
				return written, err
			}
		}
		w, err := writeNumber(dst, tok)
		written += int64(w)
		c.needsSpace = true
		return written, err
	case json.Delim:
		switch tok {
		case '[':
			c.needsSpace = false
			return c.array(dst)
		case '{':
			c.needsSpace = false
			return c.object(dst)
		}
	case bool:
		var written int64
		if c.needsSpace {
			w, err := dst.Write([]byte{' '})
			written += int64(w)
			if err != nil {
				return written, err
			}
		}
		var w int
		if tok {
			w, err = dst.Write([]byte("true"))
		} else {
			w, err = dst.Write([]byte("false"))
		}
		written += int64(w)
		c.needsSpace = true
		return written, err
	default:
		if tok == nil {
			var written int64
			if c.needsSpace {
				w, err := dst.Write([]byte{' '})
				written += int64(w)
				if err != nil {
					return written, err
				}
			}
			c.needsSpace = true
			w, err := dst.Write([]byte("null"))
			written += int64(w)
			return written, err
		}
	}
	panic(fmt.Sprintf("unknown/unexpected JSON token for value %v", tok))
}

func (c *canonicalizer) array(dst io.Writer) (int64, error) {
	var written int64
	w, err := dst.Write([]byte{'['})
	written += int64(w)
	if err != nil {
		return written, err
	}
	first := true
	for {
		if !c.dec.More() {
			_, err := c.dec.Token()
			if err != nil {
				return written, err
			}
			w, err := dst.Write([]byte{']'})
			written += int64(w)
			return written, err
		}
		if !first {
			w, err = dst.Write([]byte{','})
			written += int64(w)
			if err != nil {
				return written, err
			}
		}
		first = false
		w64, err := c.value(dst)
		c.needsSpace = false
		written += w64
		if err != nil {
			return written, err
		}
	}
}

func (c *canonicalizer) object(dst io.Writer) (int64, error) {
	var values tuples
	for {
		if !c.dec.More() {
			_, err := c.dec.Token()
			if err != nil {
				return 0, err
			}
			return c.writeObject(dst, values)
		}
		var key string
		tok, err := c.dec.Token()
		if err != nil {
			return 0, err
		}
		switch tok := tok.(type) {
		case string:
			key = tok
		default:
			return 0, fmt.Errorf("Unexpected type %T (%v) reading JSON object, expected string key", tok, tok)
		}
		buf := new(bytes.Buffer)
		_, err = c.value(buf)
		c.needsSpace = false
		if err != nil {
			return 0, err
		}
		values = append(values, tuple{key: key, val: buf.Bytes()})
	}
}

type tuple struct {
	key string
	val []byte
}
type tuples []tuple

func (t tuples) Len() int           { return len(t) }
func (t tuples) Less(i, j int) bool { return t[i].key < t[j].key }
func (t tuples) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }

func (c *canonicalizer) writeObject(dst io.Writer, values tuples) (int64, error) {
	var written int64
	w, err := dst.Write([]byte{'{'})
	written += int64(w)
	if err != nil {
		return written, err
	}
	sort.Sort(values)
	first := true
	for _, value := range values {
		if !first {
			w, err = dst.Write([]byte{','})
			written += int64(w)
			if err != nil {
				return written, err
			}
		}
		first = false
		w64, err := c.writeString(dst, value.key)
		written += w64
		if err != nil {
			return written, err
		}
		dst.Write([]byte{':'})
		w, err = dst.Write(value.val)
		written += int64(w)
		if err != nil {
			return written, err
		}
	}
	w, err = dst.Write([]byte{'}'})
	written += int64(w)
	return written, err
}

var stringExceptions = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 'b', 't', 'n', 0, 'f', 'r', 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

const hex = "0123456789abcdef"

func (c *canonicalizer) writeString(dst io.Writer, s string) (int64, error) {
	c.scratch.Reset()
	c.scratch.WriteByte('"')
	for _, r := range s {
		if r < 0x20 {
			exn := stringExceptions[r]
			if exn != 0 {
				c.scratch.WriteByte('\\')
				c.scratch.WriteByte(exn)
			} else {
				c.scratch.Write([]byte(`\u00`))
				c.scratch.WriteByte(hex[r>>4])
				c.scratch.WriteByte(hex[r&0x0f])
			}
		} else {
			if r == '\\' || r == '"' {
				c.scratch.WriteByte('\\')
			}
			c.scratch.WriteRune(r)
		}
	}
	c.scratch.WriteByte('"')
	return c.scratch.WriteTo(dst)
}

func writeNumber(dst io.Writer, f float64) (int, error) {
	if 1-(1<<53) <= f && f <= (1<<53)-1 {
		_, frac := math.Modf(f)
		if frac == 0.0 {
			return io.WriteString(dst, strconv.FormatInt(int64(f), 10))
		}
	}
	bs := strconv.AppendFloat([]byte{}, f, 'E', -1, 64)
	// this is kind of stupid, but oh well: We need to strip away pluses and
	// leading zeroes in exponents.
	var e int
	for i := range bs {
		if bs[i] == 'E' {
			e = i
		}
	}
	writeFrom := e + 1
	offset := 0
	hasPlus := bs[e+1] == '+'
	if hasPlus {
		offset++
		writeFrom = e + 2
	}
	hasLeadingZero := bs[e+2] == '0'
	if hasLeadingZero {
		offset++
		writeFrom = e + 3
	}
	for writeFrom < len(bs) {
		bs[writeFrom-offset] = bs[writeFrom]
		writeFrom++
	}
	bs = bs[:len(bs)-offset]
	return dst.Write(bs)
}
