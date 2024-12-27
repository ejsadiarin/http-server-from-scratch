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

func echoHandler(urlPath string) string {
	str := strings.Split(urlPath, "/")
	// always expect an arg in url (/echo/<arg>)
	if len(str) < 3 {
		fmt.Println("Missing string argument.")
		return "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	arg := str[2]
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(arg), arg)
	return response
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

	// custom mux
	// parse the url path
	lines := strings.Split(string(req), "\r\n")
	if len(lines) < 1 {
		fmt.Println("Malformed request")
		os.Exit(1)
	}
	requestLine := strings.Split(lines[0], " ")
	if len(requestLine) < 2 {
		fmt.Println("Malformed request")
		os.Exit(1)
	}
	urlPath := requestLine[1]
	fmt.Printf("URL Path: %s\n", urlPath)

	// switch urlPath {
	// case "/":
	//        response = "HTTP/1.1 200 OK\r\n\r\n"
	// case "/echo/...":
	//    default:
	//        response = "HTTP/1.1 404 Not Found\r\n\r\n"
	// }
	// _, err = conn.Write([]byte(response))
	// if err != nil {
	// 	fmt.Println("Error writing data", err.Error())
	// 	os.Exit(1)
	// }

	var response string
	if urlPath == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.Contains(urlPath, "/echo") {
		fmt.Println("Contains /echo")
		// echo handler logic here
		response = echoHandler(urlPath)
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	conn.Write([]byte(response))
}
