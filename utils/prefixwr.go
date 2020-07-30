package utils

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// PrefixWriter wraps an io.Writer to add a prefix on each line
// NB: caller is responsble for calling Flush().
type PrefixWriter struct {
	prefixes []string
	writer   io.Writer
	buff     bytes.Buffer
	skip     bool
}

// NewPrefixWriter creates a new PrefixWriter
func NewPrefixWriter(w io.Writer, skipFirst bool) *PrefixWriter {
	return &PrefixWriter{
		writer: w,
		skip:   skipFirst,
	}
}

// Prefix returns the prefix
func (pw *PrefixWriter) Prefix() string {
	return strings.Join(pw.prefixes, "")
}

func (pw *PrefixWriter) flush() (int, error) {
	if pw.buff.Len() > 0 {
		if pw.skip {
			pw.skip = false
		} else {
			_, err := io.WriteString(pw.writer, pw.Prefix())
			if err != nil {
				return 0, err
			}
		}

		return io.WriteString(pw.writer, pw.buff.String())
	}

	return 0, nil
}

func (pw *PrefixWriter) Write(data []byte) (int, error) {

	for i, b := range data {
		if b == '\n' {
			_, err := pw.flush()
			if err != nil {
				return i, err
			}

			_, err = fmt.Fprintln(pw.writer)
			if err != nil {
				return i, err
			}

			pw.buff.Reset()
		} else {
			pw.buff.WriteByte(b)
		}
	}

	return len(data), nil
}

// WriteString  writes a string
func (pw *PrefixWriter) WriteString(data string) (int, error) {

	for i, b := range data {
		if b == '\n' {
			_, err := pw.flush()
			if err != nil {
				return i, err
			}

			_, err = fmt.Fprintln(pw.writer)
			if err != nil {
				return i, err
			}

			pw.buff.Reset()
		} else {
			pw.buff.WriteRune(b)
		}
	}

	return len(data), nil
}

// AppendNewLineOrDie appends a new line and panics in case of an error
// here PrefixWriter will not scan the input for new lines
func (pw *PrefixWriter) AppendNewLineOrDie(s string) {
	pw.flush()

	var prefix string
	if pw.skip {
		prefix = ""
		pw.skip = false
	} else {
		prefix = pw.Prefix()
	}

	_, err := fmt.Fprintf(pw.writer, "%s%s\n", prefix, s)
	if err != nil {
		panic("AppendNewLineOrDie failed")
	}
}

// WriteOrDie writes or panics
func (pw *PrefixWriter) WriteOrDie(data []byte) {
	_, err := pw.Write(data)
	if err != nil {
		panic("WriteOrDie failed")
	}
}

// WriteStringOrDie writes a string or panics
func (pw *PrefixWriter) WriteStringOrDie(data string) {
	_, err := pw.WriteString(data)
	if err != nil {
		panic("WriteStringOrDie failed")
	}
}

// PushPrefix pushes a prefix to the stack
func (pw *PrefixWriter) PushPrefix(prefix string) {
	pw.prefixes = append(pw.prefixes, prefix)
}

// PopPrefix pops a prefix from the stack
func (pw *PrefixWriter) PopPrefix() string {
	l := len(pw.prefixes)
	if l == 0 {
		panic("popPrefix on empty prefix list")
	}
	last := pw.prefixes[l-1]
	pw.prefixes = pw.prefixes[:l-1]
	return last
}

// Flush forces PrefixWriter to flush any buffered data
func (pw *PrefixWriter) Flush() error {
	_, err := pw.flush()
	return err
}

// Done flushes buffered and and ensures prefix stack is empty
func (pw *PrefixWriter) Done() error {
	_, err := pw.flush()
	if err != nil {
		return err
	}

	if len(pw.prefixes) > 0 {
		return fmt.Errorf("Not all prefixes where popped")
	}

	return nil
}
