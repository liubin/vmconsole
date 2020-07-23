//
// Copyright (c) 2017-2018 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/go-logfmt/logfmt"
)

const (
	// Tell time.Parse() how to handle the various logfile timestamp
	// formats by providing a number of formats for the "magic" data the
	// golang time package mandates:
	//
	//     "Mon Jan 2 15:04:05 -0700 MST 2006"
	//
	dateFormat = "2006-01-02T15:04:05.999999999Z07:00"

	// The timezone of an RFC3339 timestamp can either be "Z" to denote
	// UTC, or "+/-HH:MM" to denote an actual offset.
	timezonePattern = `(` +
		`Z` +
		`|` +
		`[\+|\-]\d{2}:\d{2}` +
		`)`

	dateFormatPattern =
	// YYYY-MM-DD
	`\d{4}-\d{2}-\d{2}` +
		// time separator
		`T` +
		// HH:MM:SS
		`\d{2}:\d{2}:\d{2}` +
		// high-precision separator
		`.` +
		// nano-seconds. Note that the quantifier range is
		// required because the time.RFC3339Nano format
		// trunctates trailing zeros.
		`\d{1,9}` +
		// timezone
		timezonePattern
)

type kvPair struct {
	key   string
	value string
}

type logEntry struct {
	Msg       string `json:"msg"`
	Level     string `json:"level"`
	Ts        string `json:"ts"`
	Source    string `json:"source"`
	Version   string `json:"version"`
	Pid       string `json:"pid"`
	Subsystem string `json:"subsystem"`
	Sandbox   string `json:"sandbox"`
	Name      string `json:"name"`
}

type kvPairs []kvPair

var (
	dateFormatRE *regexp.Regexp
)

func init() {
	dateFormatRE = regexp.MustCompile(dateFormatPattern)
}

func parseLogFmtData(reader io.Reader, file string) error {
	d := logfmt.NewDecoder(reader)
	var keyvals kvPairs
	line := uint64(0)

	// A record is a single line
	for d.ScanRecord() {
		line++
		// split the line into key/value pairs
		for d.ScanKeyval() {
			key := string(d.Key())
			value := string(d.Value())
			if key == "vmconsole" && value != "" {
				pair := kvPair{
					key:   key,
					value: value,
				}

				// save key/value pair
				keyvals = append(keyvals, pair)
				break
			}

		}

		if err := d.Err(); err != nil {
			return fmt.Errorf("failed to parse file %q, line %d: %v (keyvals: %+v)",
				file, line, err, keyvals)
		}
	}

	for _, kv := range keyvals {
		var le logEntry
		err := json.Unmarshal([]byte(kv.value), &le)
		if err != nil {
			fmt.Println(kv.value)
			continue
		}
		if le.Msg == "" {
			fmt.Println(kv.value)
		} else {
			fmt.Println(le.Msg)
		}
	}

	if d.Err() != nil {
		return fmt.Errorf("failed to parse file %q line %d: %v", file, line, d.Err())
	}

	return nil
}

// parseLogFile reads a logfmt format logfile and converts it into log
// entries.
func parseLogFile(file string) error {
	// logfmt is unhappy attempting to read hex-encoded bytes in strings,
	// so hide those from it by escaping them.
	reader := NewHexByteReader(file)

	return parseLogFmtData(reader, file)
}

// parseTime attempts to convert the specified timestamp string into a Time
// object by checking it against various known timestamp formats.
func parseTime(timeString string) (time.Time, error) {
	if timeString == "" {
		return time.Time{}, errors.New("need time string")
	}

	t, err := time.Parse(dateFormat, timeString)
	if err != nil {
		return time.Time{}, err
	}

	// time.Parse() is "clever" but also doesn't check anything more
	// granular than a second, so let's be completely paranoid and check
	// via regular expression too.
	matched := dateFormatRE.FindAllStringSubmatch(timeString, -1)
	if matched == nil {
		return time.Time{},
			fmt.Errorf("expected time in format %q, got %q", dateFormatPattern, timeString)
	}

	return t, nil
}

func main() {
	if err := parseLogFile(os.Args[1]); err != nil {
		panic(err)
	}
}
