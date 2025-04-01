package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const inputFilePath = "messages.txt"

func main() {
	file, err := os.Open(inputFilePath)
	if err != nil {
		log.Fatalf("could not open %s: %s\n", inputFilePath, err)
	}
	defer file.Close()

	fmt.Printf("Reading data from %s\n", inputFilePath)
	fmt.Println("=====================================")

	buffer := make([]byte, 8, 8)
	for {
		_, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("Reading data completed.")
				os.Exit(0)
			}
			fmt.Printf("error: %s\n", err.Error())
			break
		}
		fmt.Printf("read: %s\n", buffer)
	}
}
