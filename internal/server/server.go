package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/AbdKaan/httpfromtcp/internal/request"
	"github.com/AbdKaan/httpfromtcp/internal/response"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
	handler  Handler
}

type Handler func(w *response.Writer, req *request.Request)

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func (h HandlerError) Write(w io.Writer) {
	writer := response.Writer{Writer: w}
	writer.WriteStatusLine(h.StatusCode)
	messageBytes := []byte(h.Message)
	headers := response.GetDefaultHeaders(len(messageBytes))
	writer.WriteHeaders(headers)
	w.Write(messageBytes)
}

func Serve(handler Handler, port int) (*Server, error) {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf(
			"error occured creating the listener on port %v: %v",
			port,
			err,
		)
	}
	server := &Server{listener: listener, handler: handler}
	go server.listen()
	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Server) listen() {
	for {
		// Wait for a connection
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	writer := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		writer.WriteStatusLine(response.StatusCodeBadRequest)
		body := []byte(fmt.Sprintf("Error parsing request: %v", err))
		writer.WriteHeaders(response.GetDefaultHeaders(len(body)))
		writer.WriteBody(body)
		return
	}
	s.handler(writer, req)
	return
}
