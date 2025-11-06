package runtime

import (
	"io"
	"strings"
	"sync"
)

// TemplateStream represents a streaming renderer for a template. It mirrors
// Jinja2's “Template.generate“ helper by yielding rendered fragments as they
// are produced, while still honouring the environment's trailing newline
// policy when the stream is written or collected.
type TemplateStream struct {
	chunks   chan streamChunk
	trimLast bool
	once     sync.Once
}

type streamChunk struct {
	text string
	err  error
}

func newTemplateStream(trim bool) *TemplateStream {
	return &TemplateStream{
		chunks:   make(chan streamChunk, 1),
		trimLast: trim,
	}
}

func (s *TemplateStream) emit(text string) {
	if text == "" {
		return
	}
	s.chunks <- streamChunk{text: text}
}

func (s *TemplateStream) close(err error) {
	s.once.Do(func() {
		if err != nil {
			s.chunks <- streamChunk{err: err}
		}
		close(s.chunks)
	})
}

// Next returns the next rendered fragment from the stream. When the stream is
// exhausted “io.EOF“ is returned. If rendering raised an error, that error is
// returned and the stream is closed.
func (s *TemplateStream) Next() (string, error) {
	chunk, ok := <-s.chunks
	if !ok {
		return "", io.EOF
	}
	if chunk.err != nil {
		return "", chunk.err
	}
	return chunk.text, nil
}

// Collect concatenates all remaining fragments into a single string. The
// environment's “keep_trailing_newline“ policy is honoured when producing the
// final result. Errors raised during rendering are returned to the caller.
func (s *TemplateStream) Collect() (string, error) {
	var builder strings.Builder
	for {
		chunk, err := s.Next()
		if err != nil {
			if err == io.EOF {
				result := builder.String()
				if s.trimLast {
					result = trimTrailingNewline(result)
				}
				return result, nil
			}
			return "", err
		}
		builder.WriteString(chunk)
	}
}

// WriteTo copies the remaining fragments to the supplied writer. Trailing
// newlines are trimmed to match the environment unless
// “keep_trailing_newline“ is enabled. Errors raised during rendering stop the
// stream and are returned to the caller. The number of bytes written to the
// supplied writer is returned to mirror Go's io.WriterTo contract.
func (s *TemplateStream) WriteTo(w io.Writer) (int64, error) {
	consumer := newTrimAwareWriter(w, s.trimLast)
	var written int64

	for {
		chunk, err := s.Next()
		if err != nil {
			if err == io.EOF {
				flushed, flushErr := consumer.Flush()
				written += flushed
				return written, flushErr
			}
			flushed, flushErr := consumer.Flush()
			written += flushed
			if flushErr != nil {
				return written, flushErr
			}
			return written, err
		}
		chunkWritten, writeErr := consumer.WriteChunk(chunk)
		written += chunkWritten
		if writeErr != nil {
			return written, writeErr
		}
	}
}

func trimTrailingNewline(value string) string {
	switch {
	case strings.HasSuffix(value, "\r\n"):
		return value[:len(value)-2]
	case strings.HasSuffix(value, "\n"):
		return value[:len(value)-1]
	default:
		return value
	}
}

type trimAwareWriter struct {
	writer  io.Writer
	trim    bool
	pending []byte
}

func newTrimAwareWriter(w io.Writer, trim bool) *trimAwareWriter {
	return &trimAwareWriter{writer: w, trim: trim}
}

func (w *trimAwareWriter) WriteChunk(chunk string) (int64, error) {
	var written int64

	if !w.trim {
		if len(w.pending) > 0 {
			n, err := w.writer.Write(w.pending)
			written += int64(n)
			if err != nil {
				if n > 0 {
					w.pending = w.pending[n:]
				}
				return written, err
			}
			w.pending = w.pending[:0]
		}
		if chunk == "" {
			return written, nil
		}
		n, err := io.WriteString(w.writer, chunk)
		written += int64(n)
		return written, err
	}

	if len(chunk) == 0 && len(w.pending) == 0 {
		return 0, nil
	}

	combined := append(w.pending, chunk...)
	keep := trailingNewlineLength(combined)
	flushLen := len(combined) - keep
	if flushLen > 0 {
		n, err := w.writer.Write(combined[:flushLen])
		written += int64(n)
		if err != nil {
			if n < flushLen {
				remaining := combined[n:]
				if cap(w.pending) < len(remaining) {
					w.pending = make([]byte, len(remaining))
				} else {
					w.pending = w.pending[:len(remaining)]
				}
				copy(w.pending, remaining)
			}
			return written, err
		}
	}
	if keep > 0 {
		if cap(w.pending) < keep {
			w.pending = make([]byte, keep)
		} else {
			w.pending = w.pending[:keep]
		}
		copy(w.pending, combined[len(combined)-keep:])
	} else {
		w.pending = w.pending[:0]
	}

	return written, nil
}

func (w *trimAwareWriter) Flush() (int64, error) {
	if !w.trim {
		if len(w.pending) == 0 {
			return 0, nil
		}
		n, err := w.writer.Write(w.pending)
		written := int64(n)
		if err != nil {
			if n > 0 {
				w.pending = w.pending[n:]
			}
			return written, err
		}
		w.pending = w.pending[:0]
		return written, nil
	}

	if len(w.pending) == 0 {
		return 0, nil
	}

	if trailingNewlineLength(w.pending) == len(w.pending) {
		w.pending = w.pending[:0]
		return 0, nil
	}

	n, err := w.writer.Write(w.pending)
	written := int64(n)
	if err != nil {
		if n > 0 {
			w.pending = w.pending[n:]
		}
		return written, err
	}
	w.pending = w.pending[:0]
	return written, nil
}

func trailingNewlineLength(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	last := data[len(data)-1]
	if last == '\n' {
		if len(data) >= 2 && data[len(data)-2] == '\r' {
			return 2
		}
		return 1
	}
	return 0
}

type streamWriter struct {
	stream *TemplateStream
}

func (w *streamWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	w.stream.emit(string(p))
	return len(p), nil
}
