package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilePath = "messages.txt"

func main() {
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("could not open %s: %s\n", inputFilePath, err)
	}

	lines := getLinesChannel(file)

	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}

func getLinesChannel(file io.ReadCloser) <-chan string {
	lines := make(chan string)

	fmt.Printf("Reading data from %s\n", inputFilePath)
	fmt.Println("=====================================")

	buffer := make([]byte, 8, 8)
	currentLine := ""

	go func() {
		defer file.Close()
		defer close(lines)
		for {
			n, err := file.Read(buffer)
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
