package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"strings"
)

// TODO: appropriate status codes
// TODO: use a struct-based approach

func echoHandler(urlPath string) (string, error) {
	// always expect an arg in url (/echo/<pathParamDynamic>)
	pathParamDynamic := strings.Split(urlPath, "/")[2]
	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(pathParamDynamic), pathParamDynamic)
	slog.Log(context.TODO(), slog.LevelInfo, "echo handler test")
	return response, nil
}

func userAgentHandler(req []byte) (string, error) {
	request := strings.ToLower(string(req))
	if !strings.Contains(request, "user-agent") {
		slog.Log(context.TODO(), slog.LevelError, "request doesn't include /user-agent")
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

// REQUIREMENTS:
// - Content-Type header set to application/octet-stream.
// - Content-Length header set to the size of the file, in bytes.
// - Response body set to the file contents.
func handleFile(pathParam string, directory string) (string, error) {
	filename := strings.Split(pathParam, "/")[2]
	fmt.Println("filename: ", filename)

	// if directory don't have "/" then add "/" suffix
	var filepath string
	if !(strings.HasSuffix(directory, "/")) {
		filepath = fmt.Sprintf("%s/%s", directory, filename)
	} else {
		filepath = fmt.Sprintf("%s%s", directory, filename)
	}
	fmt.Println("filepath: ", filepath)
	file, err := os.Open(filepath)
	if err != nil {
		slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("cannot open file: %s", filepath))
		return "", err
	}
	defer file.Close()

	finfo, err := file.Stat()
	if err != nil {
		slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("finfo error: %v", err))
		return "", err
	}

	contents, err := io.ReadAll(file)
	if err != nil {
		slog.Log(context.TODO(), slog.LevelError, "cannot read contents of file via io.ReadAll()")
		return "", err
	}
	fmt.Println("contents: ", string(contents))

	response := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", finfo.Size(), string(contents))

	return response, nil
}

func handleFilePost(pathParam, directory, requestBody string) (string, error) {
	filename := strings.Split(pathParam, "/")[2]
	var filepath string
	if !(strings.HasSuffix(directory, "/")) {
		filepath = fmt.Sprintf("%s/%s", directory, filename)
	} else {
		filepath = fmt.Sprintf("%s%s", directory, filename)
	}
	fmt.Println("filepath: ", filepath)

	_, err := os.Stat(filepath)
	// if file doesn't exist then return
	if os.IsExist(err) {
		slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("file already exists on directory: %s", directory))
		return "", fmt.Errorf("file already exists on directory %s: %v", directory, err)
	}

	file, err := os.Create(filepath)
	if err != nil {
		slog.Log(context.TODO(), slog.LevelError, "cannot create file")
		return "", fmt.Errorf("cannot create file: %v", err)
	}
	_, err = file.WriteString(requestBody)
	if err != nil {
		slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("cannot write string to file %s", filepath))
		return "", fmt.Errorf("cannot write string to file %s: %v", filepath, err)
	}

	response := "HTTP/1.1 201 Created\r\n\r\n"
	return response, nil
}

// Assumes length of args is >= 2 already
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

func gzipCompression(requestLines []string, response string) (string, error) {
	var encodingLine string
	for _, v := range requestLines {
		if strings.Contains(v, "Accept-Encoding:") {
			encodingLine = v
		}
	}
	fmt.Printf("encodingLine: %s\n", encodingLine)
	valSlice := strings.Split(encodingLine, " ")[1:]
	val := strings.Join(valSlice, " ")
	fmt.Printf("Accept-Encoding value (val): %s\n", val)
	// slog.Info("Accept-Encoding value (val): %s", val)

	// invalid case
	if !strings.Contains(val, "gzip") {
		slog.Error("Unsupported Accept-Encoding value. Only gzip is supported.")
		// just return response without the Content-Encoding header
		return response, nil
	}

	// if valid then append Content-Encoding: gzip to response headers
	responseSlice := strings.Split(response, "\r\n\r\n")
	responseBody := responseSlice[1]
	fmt.Println("responseBody:", responseBody)
	// compress response body with gzip
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write([]byte(responseBody))
	if err != nil {
		slog.Error("Error when compressing response body")
		return "", err
	}
	w.Close()
	responseBody = b.String()
	finalResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%v", b.Len(), responseBody)

	fmt.Println("--test gzip data--")
	r, err := gzip.NewReader(&b)
	if err != nil {
		fmt.Println("Error creating gzip reader:", err)
	}
	decompressedBody, err := io.ReadAll(r)
	if err != nil {
		fmt.Println("Error creating gzip reader:", err)
	}
	fmt.Println("Decompressed Body:", string(decompressedBody))
	fmt.Println("--test gzip data end--")

	return finalResponse, nil
}

func handleConnection(conn net.Conn) {
	// extract URL path from request
	defer conn.Close()
	req := make([]byte, 1024)
	n, err := conn.Read(req)
	if err != nil {
		slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("error reading request: %v", err))
		return
	}
	requestString := string(req[:n])

	// parse the url path
	lines := strings.Split(requestString, "\r\n")
	if len(lines) < 1 {
		slog.Log(context.TODO(), slog.LevelError, "malformed request")
		return
	}
	requestLine := strings.Split(lines[0], " ")
	if len(requestLine) < 2 {
		slog.Log(context.TODO(), slog.LevelError, "malformed request")
		return
	}
	method := requestLine[0]
	pathParam := requestLine[1]
	fmt.Printf("method: %v\n", method)
	fmt.Printf("requestString: %v\n", requestString)
	fmt.Printf("requestLine: %v\n", requestLine)
	fmt.Printf("pathParam: %s\n", pathParam)

	// custom mux
	var response string

	// gzip "middleware"
	if pathParam == "/" {
		// response = "HTTP/1.1 200 OK\r\n\r\n"
		response = "HTTP/1.1 200 OK\r\nContent-Length: 0\r\nConnection: keep-alive\r\n\r\n"
		conn.Write([]byte(response))
	}

	if strings.Contains(pathParam, "/echo") {
		fmt.Println("Contains /echo")
		response, err = echoHandler(pathParam)
		if err != nil {
			slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("error on echoHandler: %v", err))
			response = "HTTP/1.1 404 Not Found echo\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
			conn.Write([]byte(response))
			return
		}
		// gzip middleware
		response, err := gzipCompression(lines, response)
		if err != nil {
			slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("error on gzip-echoHandler: %v", err))
			response = "HTTP/1.1 404 Not Found echo\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
			conn.Write([]byte(response))
			return
		}
		conn.Write([]byte(response))
	}

	if pathParam == "/user-agent" {
		response, err = userAgentHandler(req)
		if err != nil {
			slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("error on userAgentHandler: %v", err))
			response = "HTTP/1.1 404 Not Found user-agent\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
		}
		conn.Write([]byte(response))
	}

	if strings.Contains(pathParam, "/files") && (len(os.Args) >= 2) {
		directory, exists := hasArgs()
		if exists {
			switch method {
			case "GET":
				response, err = handleFile(pathParam, directory)
				if err != nil {
					slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("error on handleFile: %v", err))
					response = "HTTP/1.1 404 Not Found files get\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
				}
			case "POST":
				requestBody := strings.Split(requestString, "\r\n\r\n")[1]
				response, err = handleFilePost(pathParam, directory, requestBody)
				if err != nil {
					slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("error on handleFilePost: %v", err))
					response = "HTTP/1.1 404 Not Found files post\r\nContent-Length: 0\r\nConnection: close\r\n\r\n"
				}
			default:
				slog.Log(context.TODO(), slog.LevelError, "Unsupported method")
				return
			}
			conn.Write([]byte(response))
		}
	}
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		slog.Log(context.TODO(), slog.LevelError, "failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			slog.Log(context.TODO(), slog.LevelError, fmt.Sprintf("error accepting connections: %v ", err))
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}
