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

		fmt.Printf("received: %#v\n", value)

		if err := writer.Write(Value{typ: "string", str: "OK"}); err != nil {
			fmt.Println("error writing to client:", err)
			return
		}
	}
}
