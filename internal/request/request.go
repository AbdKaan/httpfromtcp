package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/AbdKaan/httpfromtcp/internal/headers"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       requestState
}

type requestState int

const (
	requestStateInitialized requestState = iota
	requestStateDone
	requestStateParsingHeaders
	requestStateParsingBody
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

const crlf = "\r\n"

const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buffer := make([]byte, bufferSize, bufferSize)
	readToIndex := 0
	request := Request{
		state:   requestStateInitialized,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
	}
	for request.state != requestStateDone {
		// If the buffer is full, grow it
		if readToIndex >= len(buffer) {
			newBuffer := make([]byte, len(buffer)*2)
			copy(newBuffer, buffer)
			buffer = newBuffer
		}

		numBytesRead, err := reader.Read(buffer[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request")
				}
				break
			}
			return nil, fmt.Errorf("error reading from buffer: %v", err)
		}
		readToIndex += numBytesRead

		numBytesParsed, err := request.parse(buffer[:readToIndex])
		if err != nil {
			return nil, fmt.Errorf("error parsing data: %v", err)
		}

		// Remove the data that was successfully parsed from the buffer
		copy(buffer, buffer[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	return &request, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	requestLineText := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineText)
	if err != nil {
		return nil, 0, err
	}
	// Returns number of bytes it consumed
	return requestLine, idx + 2, nil
}

func requestLineFromString(str string) (*RequestLine, error) {
	// request-line
	// Format:
	// HTTP-version  = HTTP-name "/" DIGIT "." DIGIT
	// HTTP-name     = %s"HTTP"
	// request-line  = method SP request-target SP HTTP-version

	// Example:
	// GET /coffee HTTP/1.1

	parts := strings.Split(str, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("poorly formatted request-line: %s", str)
	}

	method := parts[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("invalid method: %s", method)
		}
	}

	requestTarget := parts[1]

	versionParts := strings.Split(parts[2], "/")
	if len(versionParts) != 2 {
		return nil, fmt.Errorf("malformed start-line: %s", str)
	}

	httpPart := versionParts[0]
	if httpPart != "HTTP" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", httpPart)
	}
	version := versionParts[1]
	if version != "1.1" {
		return nil, fmt.Errorf("unrecognized HTTP-version: %s", version)
	}

	return &RequestLine{
		Method:        method,
		RequestTarget: requestTarget,
		HttpVersion:   versionParts[1],
	}, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		bytesParsed, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, fmt.Errorf("error occured parsing headers: %v", err)
		}
		totalBytesParsed += bytesParsed
		if bytesParsed == 0 {
			break
		}
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialized:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			// something actually went wrong
			return 0, err
		}
		if n == 0 {
			// just need more data
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		bytesParsed, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return bytesParsed, nil
	case requestStateParsingBody:
		// Check if Content-Length exists
		contentLength, ok := r.Headers.Get("content-length")
		if !ok {
			r.state = requestStateDone
			return 0, nil
		}

		// Convert string Content-Length into int
		contentLengthInt, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, fmt.Errorf("malformed Content-Length: %s", err)
		}

		// Append the data into r.Body
		r.Body = append(r.Body, data...)

		// Error if Body length is greater than Content-Length
		if len(r.Body) > contentLengthInt {
			return 0, fmt.Errorf("error: body length is greater than content-length")
		}

		// Done if Body length is equal to Content-Length
		if len(r.Body) == contentLengthInt {
			r.state = requestStateDone
		}
		return len(data), nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unknown state")
	}
}
