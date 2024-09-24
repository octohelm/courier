package main_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"golang.org/x/net/http2"

	"github.com/go-courier/logr"
	"github.com/go-courier/logr/slog"
	"github.com/octohelm/courier/example/apis"
	"github.com/octohelm/courier/example/client/example"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/courierhttp/client"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	"github.com/pkg/errors"
)

var htLogger = client.HttpTransportFunc(func(req *http.Request, next client.RoundTrip) (*http.Response, error) {

	startedAt := time.Now()

	ctx, logger := logr.Start(req.Context(), "Request")
	defer logger.End()

	resp, err := next(req.WithContext(ctx))

	defer func() {
		cost := time.Since(startedAt)

		logger := logger.WithValues(
			"cost", fmt.Sprintf("%0.3fms", float64(cost/time.Millisecond)),
			"method", req.Method,
			"url", req.URL.String(),
			"metadata", req.Header,
			"content-len", req.ContentLength,
		)

		if err == nil {
			logger.WithValues("response.proto", resp.Proto).Info("success")
		} else {
			logger.Warn(errors.Wrap(err, "http request failed"))
		}
	}()

	return resp, err
})

func TestAll(t *testing.T) {
	h, err := httprouter.New(apis.R, "example")
	testingutil.Expect(t, err, testingutil.Be[error](nil))
	srv := testingutil.Serve(t, h)

	c := &example.Client{
		Endpoint:       srv.URL,
		HttpTransports: []client.HttpTransport{htLogger},
	}
	ctx := c.InjectContext(context.Background())
	ctx = logr.WithLogger(ctx, slog.Logger(slog.Default()))

	t.Run("Do Some Request", func(t *testing.T) {
		org := &example.GetOrg{
			OrgName: "test",
		}
		resp, err := example.Do(ctx, org)
		testingutil.Expect(t, err, testingutil.Be[error](nil))
		testingutil.Expect(t, resp.Name, testingutil.Be(org.OrgName))
	})

	t.Run("Do Some Request with h2", func(t *testing.T) {
		org := &example.GetOrg{
			OrgName: "test",
		}
		resp, err := example.Do(client.ContextWithRoundTripperCreator(ctx, func() http.RoundTripper {
			return &http2.Transport{
				AllowHTTP: true,
				DialTLSContext: func(ctx context.Context, network, addr string, cfg *tls.Config) (net.Conn, error) {
					return net.Dial(network, addr)
				},
			}
		}), org)
		testingutil.Expect(t, err, testingutil.Be[error](nil))
		testingutil.Expect(t, resp.Name, testingutil.Be(org.OrgName))
	})

	t.Run("Upload", func(t *testing.T) {
		v := &example.UploadBlob{
			ReadCloser: courierhttp.WrapReadCloser(bytes.NewBufferString("1234567")),
		}
		_, err := example.Do(ctx, v)
		testingutil.Expect(t, err, testingutil.Be[error](nil))
	})

	t.Run("UploadStoreBlob", func(t *testing.T) {
		v := &example.UploadStoreBlob{
			Scope:      "a/b/c",
			ReadCloser: courierhttp.WrapReadCloser(bytes.NewBufferString("1234567")),
		}
		_, err := example.Do(ctx, v)
		testingutil.Expect(t, err, testingutil.Be[error](nil))
	})

	t.Run("GetStoreBlob", func(t *testing.T) {
		v := &example.GetStoreBlob{
			Scope:  "a/b/c",
			Digest: "xxx",
		}
		resp, err := example.Do(ctx, v)
		testingutil.Expect(t, err, testingutil.Be[error](nil))
		testingutil.Expect(t, *resp, testingutil.Be("a/b/c@xxx"))
	})

	t.Run("GetFile", func(t *testing.T) {
		v := &example.GetFile{
			Path: "a/b/c",
		}
		resp, err := example.Do(ctx, v)
		testingutil.Expect(t, err, testingutil.Be[error](nil))
		testingutil.Expect(t, *resp, testingutil.Be("/a/b/c"))
	})
}
