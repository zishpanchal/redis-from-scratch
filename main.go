package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	handler := NewHandler()
	aof, err := NewAOF("database.aof")
	if err != nil {
		fmt.Println("error opening append-only file:", err)
		return
	}
	defer aof.Close()

	if err := ReplayAOF(aof, handler); err != nil {
		fmt.Println("error restoring append-only file:", err)
		return
	}
	server := NewServer(handler, aof)

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on port :6379")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection:", err)
			continue
		}

		go handleConnection(conn, server)
	}
}

func handleConnection(conn net.Conn, server *Server) {
	defer conn.Close()

	resp := NewResp(conn)
	writer := NewWriter(conn)
	for {
		value, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error reading from client:", err)
			return
		}

		result := server.Execute(value)
		if err := writer.Write(result); err != nil {
			fmt.Println("error writing to client:", err)
			return
		}
	}
}
