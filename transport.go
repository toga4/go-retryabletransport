package retryabletransport

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type ShouldRetryFunc = func(response *http.Response, err error) bool

type Transport struct {
	rt            http.RoundTripper
	shouldRetry   ShouldRetryFunc
	backoffConfig BackoffConfig
}

// Ensure at compile time that Transport implements http.RoundTripper.
var _ http.RoundTripper = &Transport{}

func NewTransport(base http.RoundTripper, shouldRetry ShouldRetryFunc, backoffConfig BackoffConfig) *Transport {
	if base == nil {
		base = http.DefaultTransport
	}

	if shouldRetry == nil {
		shouldRetry = func(*http.Response, error) bool { return false }
	}

	return &Transport{
		rt:            base,
		shouldRetry:   shouldRetry,
		backoffConfig: backoffConfig,
	}
}

// RoundTrip implements http.RoundTripper
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// req.Bodyは一度読んだら終わりでリトライできないため、事前にリクエストボディをオンメモリにバッファする
	// 効率的ではないが簡素な実装を選択した
	buf := []byte{}
	if req.Body != nil {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		buf = b
	}

	// このリクエストに対するリトライバックオフアルゴリズムを初期化
	backoff := t.backoffConfig.New()

	for {
		req.Body = ioutil.NopCloser(bytes.NewReader(buf))

		// 子の Transport の RoundTrip を実行
		res, err := t.rt.RoundTrip(req)

		// リトライするかどうかをチェックし、しない場合はそこで終了
		if !t.shouldRetry(res, err) {
			return res, err
		}

		if res != nil {
			// リトライする、かつレスポンスがある場合はここで読み捨てる
			// これは keep-alive されている TCP コネクションを再利用するために必要な処理である
			_, _ = io.Copy(ioutil.Discard, res.Body)
			res.Body.Close()
		}

		// バックオフ設定に従って一定時間待機する
		wait := backoff.Pause()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
			// リトライへ
		}
	}
}
