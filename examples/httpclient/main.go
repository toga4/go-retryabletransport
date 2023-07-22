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
	"github.com/toga4/go-retryabletransport/adapter/github.com/googleapis/gax-go.v2/gaxbackoff"
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
	_ = http.Serve(listener, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	backoffPolicy := &gaxbackoff.BackoffPolicy{
		Initial:    300 * time.Millisecond,
		Max:        2 * time.Second,
		Multiplier: 1.5,
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 20

	retryableTransport := retryabletransport.New(
		backoffPolicy,
		retryabletransport.WithTransport(transport),
		retryabletransport.WithShouldRetryError(func(r *http.Request, err error) bool {
			if errors.Is(err, syscall.ECONNRESET) {
				log.Printf("Retry HTTP request! err: %v", err)
				return true
			}
			return false
		}),
		retryabletransport.WithShouldRetryResponse(func(r *http.Response) bool {
			if r.StatusCode == http.StatusBadGateway {
				log.Printf("Retry HTTP request! status: %v", r.Status)
				return true
			}
			return false
		}),
	)

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
