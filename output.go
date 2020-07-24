package main

import (
	"fmt"
	"io"
)

type output interface {
	output(le *logEntry)
}

type consoleOutput struct {
	writer io.Writer
}

func newConsoleOutput(writer io.Writer) *consoleOutput {
	return &consoleOutput{
		writer: writer,
	}
}

func (co *consoleOutput) output(le *logEntry) {
	if le.raw != "" {
		fmt.Fprintf(co.writer, fmt.Sprintf("%s\n", le.raw))
	} else {
		// fmt.Fprintf(co.writer, fmt.Sprintf("%s\n", le.Msg))
		fmt.Printf(fmt.Sprintf("%s %s %s %s %s: %s\n", le.Ts, le.Name, le.Level, le.Source, le.Subsystem, le.Msg))
	}
}
