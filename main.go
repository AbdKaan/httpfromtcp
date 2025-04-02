package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

const port = ":42069"

func main() {
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("error listening for TCP traffic: %s\n", err.Error())
	}
	defer listener.Close()

	fmt.Println("Listening for TCP traffic on:", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("error: %s\n", err.Error())
		}
		fmt.Println("A connection has been accepted from: ", conn.RemoteAddr())

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("%s\n", line)
		}
		fmt.Println("Connection to ", conn.RemoteAddr(), "is closed")
	}
}

func getLinesChannel(rc io.ReadCloser) <-chan string {
	lines := make(chan string)

	buffer := make([]byte, 8, 8)
	currentLine := ""

	go func() {
		defer rc.Close()
		defer close(lines)
		for {
			n, err := rc.Read(buffer)
			if err != nil {
				if currentLine != "" {
					lines <- currentLine
				}
				if errors.Is(err, io.EOF) {
					return
				}
				fmt.Printf("error: %s\n", err.Error())
				os.Exit(1)
			}
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := range len(parts) - 1 {
				lines <- fmt.Sprintf("%s%s", currentLine, parts[i])
				currentLine = ""
			}
			currentLine += parts[len(parts)-1]
		}
	}()

	return lines
}
