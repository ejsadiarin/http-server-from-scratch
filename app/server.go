package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var (
	_ = net.Listen
	_ = os.Exit
)

func echoHandler(conn net.Conn, request []byte) {
	// parse the request
	requestString := string(request)
	reqSlice := strings.Split(requestString, " ")
	urlPath := reqSlice[1]
	fmt.Printf("URL Path: %s", urlPath)
	if !strings.HasPrefix(urlPath, "/echo") {
		fmt.Printf("\nInvalid URL Path: %s\n", urlPath)
		conn.Close()
		return
	}

	str := strings.Split(urlPath, "/")
	arg := str[1]
	fmt.Printf("str: %s, arg: %s\n", str, arg)
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	// extract URL path from request
	req := make([]byte, 1024)
	conn.Read(req)
	fmt.Println(string(req))
	var response string
	if !strings.HasPrefix(string(req), "GET / HTTP/1.1") {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	} else {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	}

	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing data", err.Error())
		os.Exit(1)
	}

	// if strings.Contains(string(req), "/echo") {
	// 	echoHandler(conn, req)
	// }
}
