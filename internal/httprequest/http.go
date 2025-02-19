package httprequest

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"
)

func From(r *http.Request) Request {
	return &requestInfo{
		request:    r,
		receivedAt: time.Now(),
	}
}

type requestInfo struct {
	request    *http.Request
	receivedAt time.Time
	query      url.Values
	cookies    []*http.Cookie
	params     PathValueGetter
}

func (info *requestInfo) Context() context.Context {
	return info.request.Context()
}

func (info *requestInfo) Underlying() *http.Request {
	return info.request
}

func (info *requestInfo) Method() string {
	return info.request.Method
}

func (info *requestInfo) Path() string {
	return info.request.URL.Path
}

func (info *requestInfo) Header() http.Header {
	return info.request.Header
}

func (info *requestInfo) Body() io.ReadCloser {
	if info.request.ContentLength == 0 {
		if contentType := info.request.Header.Get("Content-Type"); contentType == "" {
			if query := info.request.URL.RawQuery; len(query) > 0 {
				info.request.Header.Set("Content-Type", `application/x-www-form-urlencoded; param="value"`)

				return &readerWithHeader{
					header:     info.request.Header,
					ReadCloser: io.NopCloser(bytes.NewBufferString(query)),
				}
			}
		}
	}

	return &readerWithHeader{
		header:     info.request.Header,
		ReadCloser: info.request.Body,
	}
}

func (info *requestInfo) Value(in string, name string) string {
	values := info.Values(in, name)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (info *requestInfo) Values(in string, name string) []string {
	switch in {
	case "path":
		v := info.Param(name)
		if v == "" {
			return []string{}
		}
		p, err := url.PathUnescape(v)
		if err == nil {
			return []string{p}
		}
		return []string{v}
	case "query":
		return info.QueryValues(name)
	case "cookie":
		return info.CookieValues(name)
	case "header":
		return info.HeaderValues(name)
	}
	return []string{}
}

func (info *requestInfo) Param(name string) string {
	if info.params == nil {
		info.params = PathValueGetterFromContext(info.Context())
	}
	return info.params.PathValue(name)
}

func (info *requestInfo) QueryValues(name string) []string {
	if info.query == nil {
		info.query = info.request.URL.Query()
		// get query in form-urlencoded body
		if len(info.query) == 0 && strings.HasPrefix(info.request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			data, err := io.ReadAll(info.request.Body)
			if err == nil {
				_ = info.request.Body.Close()

				query, e := url.ParseQuery(string(data))
				if e == nil {
					info.query = query
				}

				// put back to body for custom parse
				info.request.Body = io.NopCloser(bytes.NewBuffer(data))
			}
		}
	}
	return info.query[name]
}

func (info *requestInfo) HeaderValues(name string) []string {
	if values := info.QueryValues("x-param-header-" + textproto.CanonicalMIMEHeaderKey(name)); len(values) > 0 {
		return values
	}
	return info.request.Header[textproto.CanonicalMIMEHeaderKey(name)]
}

func (info *requestInfo) CookieValues(name string) []string {
	if info.cookies == nil {
		info.cookies = info.request.Cookies()
	}

	values := make([]string, 0)
	for _, c := range info.cookies {
		if c.Name == name {
			if c.Expires.IsZero() {
				values = append(values, c.Value)
			} else if c.Expires.After(info.receivedAt) {
				values = append(values, c.Value)
			}
		}
	}
	return values
}
