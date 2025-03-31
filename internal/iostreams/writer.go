// Copyright 2022-2025 Salesforce, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package iostreams

import (
	"context"
	"io"
	"strings"

	"github.com/slackapi/slack-cli/internal/style"
)

// Writer contains implementions of io.Writer that log and output provided input
//
// Used over Printer when the Write method is needed while still wanting to have
// outputs formatted and directed to the matching stream
type Writer interface {
	// WriteOut returns the writer associated with stdout
	WriteOut() io.Writer
	// WriteErr returns the writer associated with stderr
	WriteErr() io.Writer

	// WriteDebug writes the debug message using the io.Writer implementation
	WriteDebug(context.Context) WriteDebugger

	// WriteIndent writes an indented message to the writer
	WriteIndent(io.Writer) WriteIndenter

	// WriteSecondary writes messages to the writer with secondary highlights
	WriteSecondary(io.Writer) WriteSecondarier
}

// BoundariedWriter writes output between bounds to buff and streams everything
type BoundariedWriter struct {
	Active bool
	Bounds string
	Buff   io.Writer
	Stream io.Writer
}

// Write writes the message to stream and captures between bounds in buffer
func (bw *BoundariedWriter) Write(p []byte) (n int, err error) {
	if bw.Stream != nil {
		n, err := bw.Stream.Write(p)
		if err != nil {
			return n, err
		}
	}
	if bw.Buff != nil {
		line := strings.TrimRight(string(p), EOL) // EOL removed for multiline outputs
		if strings.Contains(line, bw.Bounds) {
			for chunk := range strings.SplitSeq(line, bw.Bounds) {
				if bw.Active {
					n, err := bw.Buff.Write([]byte(chunk))
					if err != nil {
						return n, err
					}
				}
				bw.Active = !bw.Active
			}
			bw.Active = !bw.Active
		} else if bw.Active {
			n, err := bw.Buff.Write([]byte(line))
			if err != nil {
				return n, err
			}
		}
	}
	return len(p), nil
}

// BufferedWriter contains two writers to write to
type BufferedWriter struct {
	Buff   io.Writer
	Stream io.Writer
}

// Write writes the same message to both writers
func (bw BufferedWriter) Write(p []byte) (n int, err error) {
	if bw.Stream != nil {
		n, err = bw.Stream.Write(p)
		if err != nil {
			return n, err
		}
	}
	if bw.Buff != nil {
		n, err = bw.Buff.Write(p)
		if err != nil {
			return n, err
		}
	}
	return n, nil
}

// FilteredWriter writes output not between bounds
type FilteredWriter struct {
	Active bool
	Bounds string
	Stream io.Writer
}

// Write writes output not between bounds to the stream
func (fw FilteredWriter) Write(p []byte) (n int, err error) {
	if fw.Stream != nil && strings.Contains(string(p), fw.Bounds) {
		for _, chunk := range strings.Split(string(p), fw.Bounds) {
			if !fw.Active {
				n, err := fw.Stream.Write([]byte(strings.TrimSpace(chunk)))
				if err != nil {
					return n, err
				}
			}
			fw.Active = !fw.Active
		}
	}
	if fw.Stream != nil && !strings.Contains(string(p), fw.Bounds) {
		return fw.Stream.Write(p)
	}
	return len(p), nil
}

// WriteOut returns the writer associated with stdout
func (io *IOStreams) WriteOut() io.Writer {
	return io.Stdout.Writer()
}

// WriteErr returns the writer associated with stderr
func (io *IOStreams) WriteErr() io.Writer {
	return io.Stderr.Writer()
}

// WriteDebugger contains information needed to write debug logs
type WriteDebugger struct {
	ctx context.Context
	io  IOStreamer
}

// Write splits an input and writes debug logs line by line
func (wr WriteDebugger) Write(p []byte) (n int, err error) {
	lines := strings.Split(strings.TrimSpace(string(p)), "\n")
	for _, line := range lines {
		wr.io.PrintDebug(wr.ctx, line)
	}
	return len(p), nil
}

// WriteDebug writes the debug message using the io.Writer implementation
func (io *IOStreams) WriteDebug(ctx context.Context) WriteDebugger {
	return WriteDebugger{ctx: ctx, io: io}
}

// WriteIndenter contains information needed to write indented sections
type WriteIndenter struct {
	Writer io.Writer
}

// Write splits an input and writes indented lines to the writer
func (ws WriteIndenter) Write(p []byte) (n int, err error) {
	if len(p) <= 0 {
		return len(p), nil
	}
	lines := strings.Split(strings.TrimRight(string(p), "\n"), "\n")
	for _, line := range lines {
		_, err := ws.Writer.Write([]byte(style.Indent(line) + "\n"))
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// WriteIndent writes messages in an indented format
func (io *IOStreams) WriteIndent(w io.Writer) WriteIndenter {
	return WriteIndenter{Writer: w}
}

// WriteSecondarier contains information needed to write secondary text
type WriteSecondarier struct {
	Writer io.Writer
}

// Write splits an input and writes formatted lines to the writer
func (ws WriteSecondarier) Write(p []byte) (n int, err error) {
	if len(p) <= 0 {
		return len(p), nil
	}
	lines := strings.Split(strings.TrimRight(string(p), "\n"), "\n")
	for _, line := range lines {
		_, err := ws.Writer.Write([]byte(style.Secondary(line) + "\n"))
		if err != nil {
			return 0, err
		}
	}
	return len(p), nil
}

// WriteSecondary writes messages in a dim and indented format for sections
func (io *IOStreams) WriteSecondary(w io.Writer) WriteSecondarier {
	return WriteSecondarier{Writer: w}
}
