package httpd

import (
	"errors"
	"io"
	"net"
	"net/http"
	"syscall"
	"testing"
	"time"
)

type LogHttp struct {
	t       *testing.T
	httpCli *http.Client
}

func NewCli(t *testing.T) LogHttp {
	return LogHttp{t: t, httpCli: &http.Client{}}
}

func (c LogHttp) Get(url string) (resp *http.Response, body []byte, err error) {
	c.t.Helper()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return
	}

	return c.Do(req)

}

// trigger build
func (c LogHttp) Do(req *http.Request) (res *http.Response, body []byte, err error) {
	c.t.Helper()
	c.t.Logf("%s %s", req.Method, req.URL)

	// Tolerate the first few hundred ms of pod startup, where Calico CNI
	// may not have finished programming the pod's iptables/routes and the
	// initial TCP connect can fail. Retry only on connect-level errors;
	// successful HTTP responses (incl. 4xx/5xx) return immediately.
	const maxAttempts = 8
	const backoff = 250 * time.Millisecond
	for attempt := 1; ; attempt++ {
		res, err = c.httpCli.Do(req)
		if err == nil || attempt == maxAttempts || !isConnectError(err) {
			break
		}
		c.t.Logf("transient connect error (attempt %d/%d): %v", attempt, maxAttempts, err)
		time.Sleep(backoff)
	}
	if err != nil {
		return
	}

	body, err = io.ReadAll(res.Body)
	if err == nil && len(body) > 0 {
		c.t.Logf("Body: %s", body)
	}

	return
}

// isConnectError reports whether err is a transient transport-layer
// connection failure (refused, reset, network unreachable, no route),
// as distinct from a successful HTTP exchange that returned an error
// status. Returning true allows Do to retry.
func isConnectError(err error) bool {
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		return false
	}
	if sce, ok := errors.AsType[syscall.Errno](opErr.Err); ok {
		switch sce {
		case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.ENETUNREACH, syscall.EHOSTUNREACH, syscall.ETIMEDOUT:
			return true
		}
	}
	if opErr.Op == "dial" || opErr.Op == "connect" {
		return true
	}
	return false
}
