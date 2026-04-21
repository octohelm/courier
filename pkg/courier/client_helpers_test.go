package courier

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

type testClient struct {
	result Result
}

func (c testClient) Do(context.Context, any, ...Metadata) Result {
	return c.result
}

type testResult struct {
	into func(any) error
}

func (r testResult) Into(v any) (Metadata, error) {
	if r.into != nil {
		return Metadata{"X-Test": {"1"}}, r.into(v)
	}
	return nil, nil
}

type responseDataRequest struct{}

func (responseDataRequest) ResponseData() *struct{ Name string } {
	return &struct{ Name string }{}
}

type noContentRequest struct{}

func (noContentRequest) ResponseData() *NoContent {
	return &NoContent{}
}

type readCloserRequest struct{}

func (readCloserRequest) ResponseData() io.ReadCloser {
	return nil
}

func TestClientAndContextHelpers(t0 *testing.T) {
	ctx := ContextWithClient(context.Background(), "default", testClient{})

	Then(t0, "client 上下文与 DoWith 行为符合预期",
		ExpectMust(func() error {
			if ClientFromContext(ctx, "default") == nil {
				return errors.New("missing client in context")
			}
			if ClientFromContext(ctx, "missing") != nil {
				return errors.New("unexpected client")
			}
			return nil
		}),
		ExpectMust(func() error {
			c := testClient{
				result: testResult{
					into: func(v any) error {
						resp := v.(*struct{ Name string })
						resp.Name = "demo"
						return nil
					},
				},
			}
			resp, err := DoWith(context.Background(), c, responseDataRequest{})
			if err != nil {
				return err
			}
			if !reflect.DeepEqual(resp, &struct{ Name string }{Name: "demo"}) {
				return errors.New("unexpected response data")
			}
			return nil
		}),
		ExpectMust(func() error {
			c := testClient{
				result: testResult{
					into: func(v any) error {
						if v != nil {
							return errors.New("expected nil sink")
						}
						return nil
					},
				},
			}
			_, err := DoWith(context.Background(), c, noContentRequest{})
			return err
		}),
		ExpectMust(func() error {
			c := testClient{
				result: testResult{
					into: func(v any) error {
						rc := v.(*io.ReadCloser)
						*rc = io.NopCloser(strings.NewReader("stream"))
						return nil
					},
				},
			}
			resp, err := DoWith(context.Background(), c, readCloserRequest{})
			if err != nil {
				return err
			}
			defer resp.Close()

			data, err := io.ReadAll(resp)
			if err != nil {
				return err
			}
			if string(data) != "stream" {
				return errors.New("unexpected read closer response")
			}
			return nil
		}),
	)
}

func TestMetadataAndContextCompose(t0 *testing.T) {
	Then(t0, "元数据与上下文组合辅助方法符合预期",
		ExpectMust(func() error {
			meta := FromMetas(
				Metadata{"X-Trace": {"1"}},
				Metadata{"X-User": {"demo"}},
			)
			meta.Add("X-Trace", "2")
			meta.Set("X-Mode", "test")
			if meta.Get("X-Trace") != "1" || !meta.Has("X-Mode") {
				return errors.New("unexpected metadata values")
			}
			meta.Del("X-User")
			if meta.Has("X-User") {
				return errors.New("unexpected metadata delete")
			}
			if meta.String() != "X-Mode=test&X-Trace=1&X-Trace=2" {
				return errors.New("unexpected metadata string")
			}
			return nil
		}),
		ExpectMust(func() error {
			ctx := ComposeContextWith(
				func(ctx context.Context) context.Context { return context.WithValue(ctx, "a", "1") },
				func(ctx context.Context) context.Context { return context.WithValue(ctx, "b", "2") },
			)(context.Background())
			if ctx.Value("a") != "1" || ctx.Value("b") != "2" {
				return errors.New("unexpected context values")
			}
			return nil
		}),
	)
}
