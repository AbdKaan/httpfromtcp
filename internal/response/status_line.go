package response

import (
	"fmt"
	"io"
)

type StatusCode int

const (
	StatusCodeSuccess             StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case StatusCodeSuccess:
		_, err := w.Write([]byte("HTTP/1.1 200 OK\r\n"))
		if err != nil {
			return fmt.Errorf("couldn't write status line: %v", err)
		}
	case StatusCodeBadRequest:
		_, err := w.Write([]byte("HTTP/1.1 400 Bad Request\r\n"))
		if err != nil {
			return fmt.Errorf("couldn't write status line: %v", err)
		}
	case StatusCodeInternalServerError:
		_, err := w.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n"))
		if err != nil {
			return fmt.Errorf("couldn't write status line: %v", err)
		}
	default:
		_, err := w.Write(fmt.Appendf(nil, "HTTP/1.1 %d\r\n", statusCode))
		if err != nil {
			return fmt.Errorf("couldn't write status line: %v", err)
		}
	}

	return nil
}
