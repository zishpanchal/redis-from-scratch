package main

import (
	"fmt"
	"io"
	"net"
)

func main() {
	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Listening on port :6379")

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("error accepting connection:", err)
		return
	}
	defer conn.Close()

	resp := NewResp(conn)
	for {
		value, err := resp.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("error reading from client:", err)
			return
		}

		fmt.Printf("received: %#v\n", value)

		_, err = conn.Write([]byte("+OK\r\n"))
		if err != nil {
			fmt.Println("error writing to client:", err)
			return
		}
	}
}
