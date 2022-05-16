package retryabletransport

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
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("with_all_options", func(t *testing.T) {
		backoffPolicy := &testBackoffPolicy{}
		shouldRetryError := func(*http.Request, error) bool { return true }
		shouldRetryResponse := func(*http.Response) bool { return true }
		transport := http.DefaultTransport.(*http.Transport).Clone()

		tr := New(
			backoffPolicy,
			WithShouldRetryError(shouldRetryError),
			WithShouldRetryResponse(shouldRetryResponse),
			WithTransport(transport),
		)

		if tr.backoffPolicy != backoffPolicy {
			t.Errorf("expected backoffPolicy %v to be equal %v", tr.backoffPolicy, backoffPolicy)
		}
		if !tr.shouldRetryError(nil, nil) {
			t.Error("expected shouldRetryError returns true but false")
		}
		if !tr.shouldRetryResponse(nil) {
			t.Error("expected shouldRetryResponse returns true but false")
		}
		if tr.transport != transport {
			t.Errorf("expected transport %v to be equal %v", tr.transport, transport)
		}
	})

	t.Run("without_options", func(t *testing.T) {
		backoffPolicy := &testBackoffPolicy{}

		tr := New(backoffPolicy)

		if tr.backoffPolicy != backoffPolicy {
			t.Errorf("expected backoffPolicy %v to be equal %v", tr.backoffPolicy, backoffPolicy)
		}
		if tr.shouldRetryError(nil, nil) {
			t.Error("expected shouldRetryError returns false but true")
		}
		if tr.shouldRetryResponse(nil) {
			t.Error("expected shouldRetryResponse returns false but true")
		}
		if tr.transport != nil {
			t.Errorf("expected tranport is nil but %v", tr.transport)
		}
	})
}

type testBackoffPolicy struct{}
type testBackoff struct{}

func (*testBackoffPolicy) New() Backoff {
	return &testBackoff{}
}

func (*testBackoff) Pause() time.Duration {
	return 90 * time.Millisecond
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()

	backoffPolicy := &testBackoffPolicy{}

	transport := New(
		backoffPolicy,
		WithShouldRetryError(func(r *http.Request, err error) bool {
			return errors.Is(err, syscall.ECONNRESET)
		}),
		WithShouldRetryResponse(func(r *http.Response) bool {
			return r.StatusCode == http.StatusBadGateway
		}),
	)

	c := http.Client{
		Transport: transport,
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

		if atomic.LoadUint64(&calledCount) == 2 {
			t.Errorf("expected %v to be equal to %v", calledCount, 2)
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

		if atomic.LoadUint64(&calledCount) == 2 {
			t.Errorf("expected %v to be equal to %v", calledCount, 2)
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
