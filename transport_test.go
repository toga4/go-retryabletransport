package retryabletransport_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/toga4/go-retryabletransport"
)

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	base := http.DefaultTransport

	shouldRetry := func(response *http.Response, err error) bool {
		if err != nil {
			return errors.Is(err, syscall.ECONNRESET)
		}
		return response.StatusCode == http.StatusBadGateway
	}

	backoffConfig := &retryabletransport.GaxBackoffConfig{
		Initial:    25 * time.Millisecond,
		Max:        100 * time.Millisecond,
		Multiplier: 2,
	}

	c := http.Client{
		Transport: retryabletransport.NewTransport(base, shouldRetry, backoffConfig),
		Timeout:   300 * time.Millisecond,
	}

	t.Run("502_bad_gateway__deadline_exceeded", func(t *testing.T) {
		calledCount := uint64(0)
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected %v to be equal %v", r.Method, http.MethodPost)
			}

			if r.RequestURI != "/foo/bar" {
				t.Errorf("expected %v to be equal %v", r.RequestURI, "/foo/bar")
			}

			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b, []byte("hello")) {
				t.Errorf("expected %s to be equal %s", b, "hello")
			}

			w.WriteHeader(http.StatusBadGateway)
			atomic.AddUint64(&calledCount, 1)
		}))
		defer s.Close()

		body := strings.NewReader("hello")
		req, err := http.NewRequest(http.MethodPost, s.URL+"/foo/bar", body)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := c.Do(req); !os.IsTimeout(err) {
			t.Errorf("expected %v to be timeout error", err)
		}

		if atomic.LoadUint64(&calledCount) < 5 {
			t.Errorf("expected %v to be greater than %v or equal", calledCount, 5)
		}
	})

	t.Run("connection_reset__deadline_exceeded", func(t *testing.T) {
		l, err := net.Listen("tcp", ":8080")
		if err != nil {
			t.Fatal(err)
		}
		defer l.Close()

		calledCount := uint64(0)
		go func() {
			for {
				conn, err := l.Accept()
				if err != nil {
					if strings.Contains(err.Error(), "use of closed network connection") {
						return
					}
					panic(err)
				}
				data := make([]byte, 1)
				if _, err := conn.Read(data); err != nil {
					panic(err)
				}

				if err := conn.Close(); err != nil {
					panic(err)
				}

				atomic.AddUint64(&calledCount, 1)
			}
		}()

		body := strings.NewReader("hello")
		url := fmt.Sprintf("http://%s/foo/bar", "localhost:8080")
		req, err := http.NewRequest(http.MethodPost, url, body)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := c.Do(req); !os.IsTimeout(err) {
			t.Errorf("expected %v to be timeout error", err)
		}

		if atomic.LoadUint64(&calledCount) < 5 {
			t.Errorf("expected %v to be greater than %v or equal", calledCount, 5)
		}
	})

	t.Run("502_bad_gateway__successful", func(t *testing.T) {
		calledCount := uint64(0)
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected %v to be equal %v", r.Method, http.MethodPost)
			}

			if r.RequestURI != "/foo/bar" {
				t.Errorf("expected %v to be equal %v", r.RequestURI, "/foo/bar")
			}

			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b, []byte("hello")) {
				t.Errorf("expected %s to be equal %s", b, "hello")
			}

			if atomic.LoadUint64(&calledCount) < 3 {
				w.WriteHeader(http.StatusBadGateway)
			} else {
				if _, err := io.WriteString(w, "world"); err != nil {
					t.Fatal(err)
				}
			}
			atomic.AddUint64(&calledCount, 1)
		}))
		defer s.Close()

		body := strings.NewReader("hello")
		req, err := http.NewRequest(http.MethodPost, s.URL+"/foo/bar", body)
		if err != nil {
			t.Fatal(err)
		}

		res, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(b, []byte("world")) {
			t.Errorf("expected %s to be equal %s", b, "world")
		}
	})
}
