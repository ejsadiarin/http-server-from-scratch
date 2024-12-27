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

func echoHandler(urlPath string) (string, error) {
	str := strings.Split(urlPath, "/")
	// always expect an arg in url (/echo/<arg>)
	if len(str) < 3 {
		fmt.Println("Missing string argument.")
		return "HTTP/1.1 404 Not Found\r\n\r\n", fmt.Errorf("Missing string argument")
	}
	arg := str[2]
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(arg), arg)
	return response, nil
}

func userAgentHandler(req []byte) (string, error) {
	if !strings.Contains(string(req), "User-Agent") {
		// return "HTTP/1.1 404 Not Found\r\n\r\n"
		return "HTTP/1.1 404 Not Found\r\n\r\n User-Agent not found\r\n", fmt.Errorf("No User-Agent header found.")
	}
	lines := strings.Split(string(req), "\r\n")
	fmt.Printf("Lines (len %d): %v\n", len(lines), lines)
	for i, line := range lines {
		fmt.Printf("%d: %s\n", i, line)
	}
	var idx int
	for i, v := range lines {
		if strings.Contains(v, "User-Agent") {
			idx = i
		}
	}
	UALine := lines[idx]
	UASlice := strings.Split(UALine, ":")
	UAValue := strings.TrimSuffix(UASlice[1], "\r\n")
	UAValue = strings.TrimSpace(UAValue)
	fmt.Println(UAValue)
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\n\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(UAValue), UAValue)
	return response, nil
}

func handleConnection(conn net.Conn) {
	// extract URL path from request
	defer conn.Close()
	req := make([]byte, 1024)
	conn.Read(req)
	fmt.Println(string(req))

	// parse the url path
	lines := strings.Split(string(req), "\r\n")
	if len(lines) < 1 {
		fmt.Println("Malformed request")
		return
	}
	requestLine := strings.Split(lines[0], " ")
	if len(requestLine) < 2 {
		fmt.Println("Malformed request")
		return
	}
	urlPath := requestLine[1]
	fmt.Printf("URL Path: %s\n", urlPath)

	// custom mux
	var response string
	if urlPath == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if strings.Contains(urlPath, "/echo") {
		fmt.Println("Contains /echo")
		// echo handler logic here
		response, _ = echoHandler(urlPath)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
	} else if urlPath == "/user-agent" {
		response, _ = userAgentHandler(req)
		// if err != nil {
		// 	fmt.Println(err)
		// 	return
		// }
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}

	conn.Write([]byte(response))
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

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
