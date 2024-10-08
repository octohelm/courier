package internal

import (
	"context"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"slices"
	"sort"
	"sync"

	"github.com/octohelm/courier/internal/httprequest"
	"github.com/octohelm/courier/internal/jsonflags"
	"github.com/octohelm/courier/internal/pathpattern"
	validatorerrors "github.com/octohelm/courier/pkg/validator/errors"
)

type Request struct {
	ParamValue
}

var locations = slices.Values([]string{"header", "query", "path", "cookie", "body"})

func (p *Request) MarshalRequest(ctx context.Context, method string, path pathpattern.Segments) (*http.Request, error) {
	s, err := jsonflags.Structs.StructFields(p.Type())
	if err != nil {
		return nil, err
	}

	var body io.ReadCloser

	query := url.Values{}
	headers := http.Header{}
	pathParams := map[string]string{}

	for sf := range s.LocatedStructField("body") {
		rv := sf.GetOrNewAt(p.Value)

		cw, err := New(sf.Type, sf.Tag.Get("mime"), "marshal")
		if err != nil {
			return nil, err
		}

		body, err = AsReadCloser(ctx, cw, rv, headers)
		if err != nil {
			return nil, err
		}

		// only one requestBody
		break
	}

	for sf := range s.LocatedStructField("path") {
		values, err := p.MarshalValues(ctx, sf)
		if err != nil {
			return nil, err
		}

		if len(values) > 0 {
			pathParams[sf.Name] = values[0]
		}
	}

	for sf := range s.LocatedStructField("query") {
		values, err := p.MarshalValues(ctx, sf)
		if err != nil {
			return nil, err
		}

		if values != nil {
			query[sf.Name] = values
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, path.Encode(pathParams), body)
	if err != nil {
		return nil, err
	}

	if len(query) > 0 {
		req.URL.RawQuery = query.Encode()
	}

	for sf := range s.LocatedStructField("header") {
		values, err := p.MarshalValues(ctx, sf)
		if err != nil {
			return nil, err
		}

		if values != nil {
			req.Header[textproto.CanonicalMIMEHeaderKey(sf.Name)] = values
		}
	}

	cookies := url.Values{}

	for sf := range s.LocatedStructField("cookie") {
		values, err := p.MarshalValues(ctx, sf)
		if err != nil {
			return nil, err
		}

		if len(values) > 0 {
			cookies[sf.Name] = values
		}
	}

	if n := len(cookies); n > 0 {
		names := make([]string, n)
		i := 0
		for name := range cookies {
			names[i] = name
			i++
		}
		sort.Strings(names)

		for _, name := range names {
			values := cookies[name]

			for i := range values {
				req.AddCookie(&http.Cookie{
					Name:  name,
					Value: values[i],
				})
			}
		}
	}

	for k, vv := range headers {
		req.Header[k] = vv
	}

	return req, err
}

func (p *Request) UnmarshalRequest(req *http.Request) error {
	return p.UnmarshalRequestInfo(httprequest.From(req))
}

func (p *Request) UnmarshalRequestInfo(r httprequest.Request) error {
	s, err := jsonflags.Structs.StructFields(p.Type())
	if err != nil {
		return err
	}

	var errs []error

	once := &sync.Once{}

	for loc := range locations {
		for sf := range s.LocatedStructField(loc) {
			if loc == "body" {
				once.Do(func() {
					if err := p.unmarshalBody(sf, r); err != nil {
						errs = append(errs, validatorerrors.WrapLocation(err, loc))
					}
				})
			} else {
				if err := p.UnmarshalValues(r.Context(), sf, r.Values(loc, sf.Name)); err != nil {
					errs = append(errs, validatorerrors.WrapLocation(err, loc))
				}
			}
		}
	}

	return validatorerrors.Join(errs...)
}

func (p *Request) unmarshalBody(sf *jsonflags.StructField, request httprequest.Request) error {
	body := request.Body()
	if body == nil {
		return nil
	}

	rv := sf.GetOrNewAt(p.Value)

	t, err := New(sf.Type, sf.Tag.Get("mime"), "unmarshal")
	if err != nil {
		return err
	}

	return t.ReadAs(request.Context(), body, rv.Addr())
}
