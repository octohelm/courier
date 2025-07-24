package main_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"testing"

	"github.com/go-courier/logr"
	"github.com/go-courier/logr/slog"
	"github.com/octohelm/courier/example/apis"
	"github.com/octohelm/courier/example/client/example"
	domainorg "github.com/octohelm/courier/example/pkg/domain/org"
	"github.com/octohelm/courier/internal/testingutil"
	"github.com/octohelm/courier/pkg/courierhttp/handler/httprouter"
	testingx "github.com/octohelm/x/testing"
	"github.com/octohelm/x/testing/bdd"
)

func TestAll(t *testing.T) {
	h := bdd.Must(httprouter.New(apis.R, "example"))

	hh := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		raw, _ := httputil.DumpRequest(request, true)
		fmt.Println(string(raw))

		h.ServeHTTP(writer, request)
	})

	for i, srv := range []*httptest.Server{
		testingutil.ServeWithH2C(t, hh),
		testingutil.Serve(t, hh),
	} {
		t.Run(fmt.Sprintf("serve http/%d", 2-i), func(t *testing.T) {
			c := &example.Client{
				Endpoint:   srv.URL,
				SupportH2C: i == 0,
			}

			ctx := c.InjectContext(context.Background())
			ctx = logr.WithLogger(ctx, slog.Logger(slog.Default()))

			t.Run("Do Some Request", func(t *testing.T) {
				org := &example.GetOrg{}
				org.OrgName = "test"

				resp, err := example.Do(ctx, org)
				testingx.Expect(t, err, testingx.BeNil[error]())

				testingx.Expect(t, resp.Name, testingx.Be(org.OrgName))
				testingx.Expect(t, resp.Type, testingx.Be(domainorg.TYPE__GOV))
			})

			t.Run("Upload", func(t *testing.T) {
				v := &example.UploadBlob{}
				v.RequestBody = io.NopCloser(bytes.NewBufferString("1234567"))

				_, err := example.Do(ctx, v)
				testingx.Expect(t, err, testingx.BeNil[error]())
			})

			t.Run("UploadStoreBlob", func(t *testing.T) {
				v := &example.UploadStoreBlob{}
				v.Scope = "a/b/c"
				v.RequestBody = io.NopCloser(bytes.NewBufferString("1234567"))

				_, err := example.Do(ctx, v)
				testingx.Expect(t, err, testingx.BeNil[error]())
			})

			t.Run("GetStoreBlob", func(t *testing.T) {
				v := &example.GetStoreBlob{}
				v.Scope = "a/b/c"
				v.Digest = "xxx"

				resp, err := example.Do(ctx, v)
				testingx.Expect(t, err, testingx.BeNil[error]())
				testingx.Expect(t, *resp, testingx.Be("a/b/c@xxx"))
			})

			t.Run("GetFile", func(t *testing.T) {
				v := &example.GetFile{}
				v.Path = "a/b/c"

				resp, err := example.Do(ctx, v)
				testingx.Expect(t, err, testingx.BeNil[error]())
				testingx.Expect(t, *resp, testingx.Be("a/b/c"))
			})
		})
	}
}
