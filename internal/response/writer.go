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

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}

	writtenBytesTotal := 0

	n, err := fmt.Fprintf(w.Writer, "%x\r\n", len(p))
	if err != nil {
		return writtenBytesTotal, fmt.Errorf("couldn't write chunked body: %v", err)
	}
	writtenBytesTotal += n

	n, err = w.Writer.Write(fmt.Appendf(p, "\r\n"))
	writtenBytesTotal += n

	return writtenBytesTotal, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != writerStateBody {
		return 0, fmt.Errorf("cannot write body in state %d", w.writerState)
	}
	n, err := w.Writer.Write([]byte("0\r\n\r\n"))
	return n, err
}
