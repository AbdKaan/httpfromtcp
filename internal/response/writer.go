package response

import (
	"fmt"
	"io"

	"github.com/AbdKaan/httpfromtcp/internal/headers"
)

type Writer struct {
	Writer      io.Writer
	writerState writerState
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{Writer: w, writerState: writerStateStatusLine}
}

type writerState int

const (
	writerStateStatusLine writerState = iota
	writerStateHeaders
	writerStateBody
)

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != writerStateStatusLine {
		return fmt.Errorf("cannot write status line in state %d", w.writerState)
	}

	defer func() {
		w.writerState = writerStateHeaders
	}()

	_, err := w.Writer.Write(getStatusLine(statusCode))

	return err
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != writerStateHeaders {
		return fmt.Errorf("cannot write headers in state %d", w.writerState)
	}

	defer func() {
		w.writerState = writerStateBody
	}()

	for key, value := range headers {
		_, err := w.Writer.Write(fmt.Appendf(nil, "%s: %s\r\n", key, value))
		if err != nil {
			return fmt.Errorf("couldn't write headers: %v", err)
		}
	}
	_, err := w.Writer.Write([]byte("\r\n"))
	return err
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	return w.Writer.Write(p)
}
