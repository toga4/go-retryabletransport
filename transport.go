package retryabletransport

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

type ShouldRetryFunc = func(response *http.Response, err error) bool

type Transport struct {
	rt          http.RoundTripper
	shouldRetry ShouldRetryFunc
	backoff     Backoff
}

// Ensure at compile time that RetryableTransport implements http.RoundTripper.
var _ http.RoundTripper = &Transport{}

func NewTransport(base http.RoundTripper, shouldRetry ShouldRetryFunc, backoff Backoff) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	if shouldRetry == nil {
		shouldRetry = func(*http.Response, error) bool { return false }
	}

	return &Transport{
		rt:          base,
		shouldRetry: shouldRetry,
		backoff:     backoff,
	}
}

// RoundTrip implements http.RoundTripper
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// req.Bodyは一度読んだら終わりでリトライできないため、事前にリクエストボディをオンメモリにバッファする
	// 効率的ではないが簡素な実装を選択した
	buf := []byte{}
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		buf = b
	}

	for {
		req.Body = io.NopCloser(bytes.NewReader(buf))

		// 子の Transport の RoundTrip を実行
		res, err := t.rt.RoundTrip(req)

		// リトライするかどうかをチェックし、しない場合はそこで終了
		if !t.shouldRetry(res, err) {
			return res, err
		}

		if res != nil {
			// リトライする、かつレスポンスがある場合はここで読み捨てる
			// これは keep-alive されている TCP コネクションを再利用するために必要な処理である
			_, _ = io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}

		// backoff 設定に従って一定時間待機する
		if err := t.wait(req.Context()); err != nil {
			return nil, err
		}
	}
}

func (t *Transport) wait(ctx context.Context) error {
	wait := t.backoff.Pause()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait):
		return nil
	}
}
