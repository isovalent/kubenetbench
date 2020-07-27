package utils

import (
	"bytes"
	"fmt"
	"io"
)

// PrefixWriter wraps an io.Writer to add a prefix on each line
// NB: caller is responsble for calling Flush().
type PrefixWriter struct {
	prefix string
	writer io.Writer
	buff   bytes.Buffer
	skip   bool
}

// NewPrefixWriter creates a new PrefixWriter
func NewPrefixWriter(w io.Writer, p string, skipFirst bool) *PrefixWriter {
	return &PrefixWriter{
		prefix: p,
		writer: w,
		skip:   skipFirst,
	}
}

func (pw *PrefixWriter) flush() (int, error) {
	if pw.buff.Len() > 0 {
		if pw.skip {
			pw.skip = false
		} else {
			_, err := io.WriteString(pw.writer, pw.prefix)
			if err != nil {
				return 0, err
			}
		}
	}

	return io.WriteString(pw.writer, pw.buff.String())
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

// Flush forces PrefixWriter to flush any buffered data
func (pw *PrefixWriter) Flush() error {
	_, err := pw.flush()
	return err
}
