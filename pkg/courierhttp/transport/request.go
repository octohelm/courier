package transport

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/octohelm/courier/pkg/courierhttp/handler"

	"github.com/octohelm/courier/pkg/courierhttp"
)

func FromHttpRequest(r *http.Request, service string) courierhttp.Request {
	return &requestInfo{
		service:    service,
		request:    r,
		receivedAt: time.Now(),
	}
}

type requestInfo struct {
	service    string
	request    *http.Request
	receivedAt time.Time
	query      url.Values
	cookies    []*http.Cookie
	params     handler.ParamGetter
}

func (info *requestInfo) ServiceName() string {
	return info.service
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

func (info *requestInfo) Context() context.Context {
	return info.request.Context()
}

func (info *requestInfo) Body() io.ReadCloser {
	return info.request.Body
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
		info.params = handler.ParamGetterFromContext(info.Context())
	}
	return info.params.ByName(name)
}

func (info *requestInfo) QueryValues(name string) []string {
	if info.query == nil {
		info.query = info.request.URL.Query()

		if info.request.Method == http.MethodGet && len(info.query) == 0 && info.request.ContentLength > 0 {
			// get query in form-urlencoded body
			if strings.HasPrefix(info.request.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
				data, err := ioutil.ReadAll(info.request.Body)
				if err == nil {
					info.request.Body.Close()

					query, e := url.ParseQuery(string(data))
					if e == nil {
						info.query = query
					}
				}
			}
		}
	}
	return info.query[name]
}

func (info *requestInfo) HeaderValues(name string) []string {
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
