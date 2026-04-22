package operatortest

import (
	"context"
	"net/http"
	"testing"

	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/client"
)

type serverCtxKey struct{}

type testPing struct {
	courierhttp.MethodGet `path:"/ping"`

	Trace string `name:"X-Trace,omitzero" in:"header"`
}

type pingResponse struct {
	Trace   string `json:"trace"`
	Context string `json:"context"`
}

func (req *testPing) Output(ctx context.Context) (any, error) {
	return &pingResponse{
		Trace:   req.Trace,
		Context: ctx.Value(serverCtxKey{}).(string),
	}, nil
}

func TestServerHelpers(t *testing.T) {
	s := Serve(context.WithValue(context.Background(), serverCtxKey{}, "from-server"), &testPing{})
	defer s.Close()

	seenTransport := 0
	s.ApplyHttpTransport(client.HttpTransportFunc(func(req *http.Request, next client.RoundTrip) (*http.Response, error) {
		seenTransport++
		req.Header.Set("X-Trace", "trace-1")
		return next(req)
	}))

	result := s.Do(context.Background(), &testPing{})

	var out pingResponse
	if _, err := result.Into(&out); err != nil {
		t.Fatalf("unexpected request error: %v", err)
	}
	if seenTransport != 1 {
		t.Fatalf("expected transport to run once, got %d", seenTransport)
	}
	if out.Trace != "trace-1" {
		t.Fatalf("unexpected trace value: %q", out.Trace)
	}
	if out.Context != "from-server" {
		t.Fatalf("unexpected context value: %q", out.Context)
	}
}
