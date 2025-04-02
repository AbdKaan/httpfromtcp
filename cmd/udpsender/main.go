package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

const serverAddr = "localhost:42069"

func main() {
	udpadr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving UDP address: %s\n", err)
		os.Exit(1)
	}

	udpConn, err := net.DialUDP("udp", nil, udpadr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error dialing up UDP connection: %s\n", err)
		os.Exit(1)
	}
	defer udpConn.Close()

	fmt.Printf(
		"Sending to %s. Type your message and press Enter to send. Press Ctrl+C to exit.\n",
		serverAddr,
	)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf(">")
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
			os.Exit(1)
		}

		_, err = udpConn.Write([]byte(message))
		if err != nil {
			fmt.Fprintf(os.Stderr, "error sending message: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Message sent: %s", message)
	}
}
