package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/toga4/go-retryabletransport"
)

func Server() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	// connection reset by peer x 3
	for i := 0; i < 3; i++ {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
		data := make([]byte, 1)
		if _, err := conn.Read(data); err != nil {
			log.Fatal(err)
		}

		if err := conn.Close(); err != nil {
			log.Println(err)
		}
	}

	// and then, response
	i := 0
	http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if i == 0 {
			// 502 BadGateway
			i++
			w.WriteHeader(http.StatusBadGateway)
		} else {
			// 200 OK
			fmt.Fprintf(w, "hello")
		}
	}))
}

func ExecRequest() {

	httpTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 20 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	shouldRetry := func(res *http.Response, err error) bool {
		switch {
		case errors.Is(err, syscall.ECONNRESET):
			log.Printf("Retry HTTP request! err: %v", err)
			return true
		case err != nil:
			return false
		case res.StatusCode == http.StatusBadGateway:
			log.Printf("Retry HTTP request! status: %v", res.Status)
			return true
		default:
			return false
		}
	}
	backoffConfig := &retryabletransport.GaxBackoffConfig{
		Initial:    500 * time.Millisecond,
		Max:        32 * time.Second,
		Multiplier: 1.5,
	}

	retryableTransport := retryabletransport.NewTransport(httpTransport, shouldRetry, backoffConfig)

	client := http.Client{
		Timeout:   5 * time.Second,
		Transport: retryableTransport,
	}

	reqBody := strings.NewReader("test body")
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/", reqBody)
	if err != nil {
		log.Printf("NewRequest err: %v", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Do err: %v", err)
		return
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read err: %v", err)
		return
	}

	log.Printf("Got Response: %s", b)
}

func main() {
	go Server()
	ExecRequest()
}
