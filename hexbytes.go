//
// Copyright (c) 2017-2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

// FROM https://github.com/kata-containers/tests/blob/master/cmd/log-parser/hexbytes.go
package main

import (
	"bufio"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// HexByteReader is an I/O Reader type.
type HexByteReader struct {
	file    *os.File
	reader  io.Reader
	data    []byte
	scanner *bufio.Scanner

	buffer []byte

	// total length of "data"
	len int

	// how much of "data" has been sent back to the caller
	offset int
}

func NewHexByteFileReader(file *os.File) *HexByteReader {
	return &HexByteReader{
		file: file,
	}
}
func NewHexByteStreamReader(reader io.Reader) *HexByteReader {
	scanner := bufio.NewScanner(reader)

	return &HexByteReader{
		reader:  reader,
		scanner: scanner,
	}
}

func (r *HexByteReader) Read(p []byte) (n int, err error) {
	if r.file != nil {
		return r.fileReader(p)
	} else {
		return r.streamReader(p)
	}
}

func copy(s, d []byte, size int) {
	for i := 0; i < size; i++ {
		d[i] = s[i]
	}
}

func (r *HexByteReader) streamReader(p []byte) (n int, err error) {
	eof := false
	for {
		bufferedSize := len(r.buffer)
		if bufferedSize > 0 {
			writeSize := len(p)
			size := 0
			if bufferedSize >= writeSize {
				size = writeSize
			} else {
				size = bufferedSize
			}

			copy(r.buffer, p, size)

			if size == bufferedSize {
				r.buffer = []byte{}
			} else {
				r.buffer = r.buffer[(size + 1):]
			}

			return size, nil
		}

		if eof {
			return 0, io.EOF
		}

		if r.scanner.Scan() {
			line := r.scanner.Text()
			line = replace(line)
			r.buffer = append(r.buffer, []byte(line)...)
			r.buffer = append(r.buffer, 0xa)
		} else {
			eof = true
			return
		}
	}
}

func replace(s string) string {
	s = strings.Replace(s, `\x`, `\\x`, -1)
	// restore if the old string is `\\x`
	s = strings.Replace(s, `\\\x`, `\\x`, -1)
	return s
}

// Read is a Reader that converts "\x" to "\\x"
func (r *HexByteReader) fileReader(p []byte) (n int, err error) {
	size := len(p)

	if r.data == nil {
		// read the entire file
		bytes, err := ioutil.ReadAll(r.file)
		if err != nil {
			return 0, err
		}

		// although logfmt is happy to parse an empty file, this is
		// surprising to users, so make it an error.
		if len(bytes) == 0 {
			return 0, errors.New("file is empty")
		}

		// perform the conversion
		s := string(bytes)
		result := strings.Replace(s, `\x`, `\\x`, -1)

		// store the data
		r.data = []byte(result)
		r.len = len(r.data)
		r.offset = 0
	}

	// calculate how much data is left to copy
	remaining := r.len - r.offset

	if remaining == 0 {
		return 0, io.EOF
	}

	// see how much data can be copied on this call
	limit := size

	if remaining < limit {
		limit = remaining
	}

	for i := 0; i < limit; i++ {
		// index into the stored data
		src := r.offset

		// copy
		p[i] = r.data[src]

		// update
		r.offset++
	}

	return limit, nil
}
