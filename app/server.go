package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func echoHandler(urlPath string) (string, error) {
	// always expect an arg in url (/echo/<pathParamDynamic>)
	pathParamDynamic := strings.Split(urlPath, "/")[2]
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathParamDynamic), pathParamDynamic)
	return response, nil
}

func userAgentHandler(req []byte) (string, error) {
	request := strings.ToLower(string(req))
	if !strings.Contains(request, "user-agent") {
		return "HTTP/1.1 404 Not Found\r\n\r\n", fmt.Errorf("no user-agent header found")
	}
	fmt.Println(request)
	lines := strings.Split(request, "\r\n")
	var val string
	for _, l := range lines {
		if strings.Contains(l, "user-agent:") {
			val = strings.TrimSpace(strings.Split(l, ":")[1])
			fmt.Println("val:", val)
			break
		}
	}
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(val), val)
	fmt.Println(response)
	return response, nil
}

func handleFile(pathParam string, directory string) (string, error) {
	filename := strings.Split(pathParam, "/")[2]
	// TODO: response:
	// - Content-Type header set to application/octet-stream.
	// - Content-Length header set to the size of the file, in bytes.
	// - Response body set to the file contents.
	fmt.Println("filename: ", filename)

	// if directory don't have "/" then add "/" suffix
	var filepath string
	if !(strings.HasSuffix(directory, "/")) {
		// add "/" suffix
		filepath = fmt.Sprintf("%s/%s", directory, filename)
	} else {
		filepath = fmt.Sprintf("%s%s", directory, filename)
	}
	fmt.Println("filepath: ", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		return "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", err
	}
	defer file.Close()

	finfo, err := file.Stat()
	if err != nil {
		return "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", err
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		return "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", err
	}
	fmt.Println("contents: ", string(contents))

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", finfo.Size(), string(contents))

	return response, nil
}

func hasArgs() (string, bool) {
	flag := os.Args[1]
	switch flag {
	case "--directory":
		d := os.Args[2]
		return d, true
	default:
		return "", false
	}
}

func handleConnection(conn net.Conn) {
	// extract URL path from request
	defer conn.Close()
	req := make([]byte, 1024)
	n, err := conn.Read(req)
	if err != nil {
		fmt.Println("Error reading request:", err)
		return
	}
	requestString := string(req[:n])

	// parse the url path
	lines := strings.Split(requestString, "\r\n")
	if len(lines) < 1 {
		fmt.Println("Malformed request")
		return
	}
	requestLine := strings.Split(lines[0], " ")
	if len(requestLine) < 2 {
		fmt.Println("Malformed request")
		return
	}
	pathParam := requestLine[1]
	fmt.Printf("requestString: %v\n", requestString)
	fmt.Printf("requestLine: %v\n", requestLine)
	fmt.Printf("pathParam: %s\n", pathParam)

	// custom mux
	var response string

	if pathParam == "/" {
		// response = "HTTP/1.1 200 OK\r\n\r\n"
		response = "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: keep-alive\r\n\r\n"
	}

	if strings.Contains(pathParam, "/echo") {
		fmt.Println("Contains /echo")
		response, err = echoHandler(pathParam)
		if err != nil {
			fmt.Println(err)
			response = "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
		}
		conn.Write([]byte(response))
	}

	if pathParam == "/user-agent" {
		response, err = userAgentHandler(req)
		if err != nil {
			fmt.Println(err)
			response = "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
		}
		conn.Write([]byte(response))
	}

	if strings.Contains(pathParam, "/files") && (len(os.Args) >= 2) {
		directory, exists := hasArgs()
		if exists {
			response, err = handleFile(pathParam, directory)
			if err != nil {
				fmt.Println(err)
				response = "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
			}
			conn.Write([]byte(response))
		}
	}

	response = "HTTP/1.1 404 Not Found\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
	_, err = conn.Write([]byte(response))
	if err != nil {
		fmt.Println("Error writing response:", err)
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
