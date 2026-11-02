package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// Experimento goroutine (paralelo vs serie):
//  1. Agregá "time" a los imports y meté time.Sleep(3 * time.Second) al inicio de handleConnection.
//  2. Terminal 1 (servidor): go run ./tcp
//  3. Terminal 2 (2 clientes a la vez): time ( go run ./tcp/client & go run ./tcp/client & wait )
//     - CON "go handleConnection(conn)" -> real ~3s (paralelo).
//     - SIN el "go" (handleConnection(conn) directo) -> real ~6s (en serie, el 2do espera al 1ro).
//
// Al terminar: sacá el Sleep y dejá el "go" delante de handleConnection.

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	for {
		// Read and write interfaces are very important
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Create a new reader from the connection
	reader := bufio.NewReader(conn)
	// Read the command line from the client
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("error reading command: %v\n", err)
		return
	}

	// Trim the newline character and split the line into command and resource
	parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
	if len(parts) != 2 {
		fmt.Fprintf(conn, "Invalid command format. "+
			"Expected format: COMMAND:RESOURCE\n")
		return
	}

	command := parts[0]
	resource := parts[1]
	log.Printf("Received command: %s %s\n", command, resource)

	time.Sleep(3 * time.Second)

	// Handle the command
	switch command {
	case "GET":
		handleGet(conn, resource)
	default:
		fmt.Printf("Unknown command: %s\n", command)
	}
}

func handleGet(conn net.Conn, resource string) {
	fmt.Fprintf(conn, "GET command received for resource %s\n", resource)
}
