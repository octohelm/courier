package transport

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"sync"

	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/statuserror"
	"github.com/octohelm/courier/pkg/transformer"
	"github.com/octohelm/courier/pkg/transformer/core"
	"github.com/octohelm/courier/pkg/validator"
	reflectx "github.com/octohelm/x/reflect"
	typex "github.com/octohelm/x/types"
	"github.com/pkg/errors"
)

type IncomingTransport interface {
	UnmarshalOperator(ctx context.Context, info courierhttp.Request, op any) error
	WriteResponse(ctx context.Context, rw http.ResponseWriter, result any, info courierhttp.Request)
}

var incomingTransports = sync.Map{}

func NewIncomingTransport(ctx context.Context, v any) (IncomingTransport, error) {
	typ := reflectx.Deref(reflect.TypeOf(v))

	if v, ok := incomingTransports.Load(typ); ok {
		return v.(IncomingTransport), nil
	}

	t := &incomingTransport{}

	t.InParameters = map[string][]core.RequestParameter{}
	t.Type = typ

	err := core.EachRequestParameter(ctx, typex.FromRType(t.Type), func(rp *core.RequestParameter) {
		if rp.In == "" {
			return
		}
		t.InParameters[rp.In] = append(t.InParameters[rp.In], *rp)
	})
	if err != nil {
		return nil, err
	}

	incomingTransports.Store(typ, t)

	return t, nil
}

type incomingTransport struct {
	Type         reflect.Type
	InParameters map[string][]transformer.RequestParameter
}

func (i *incomingTransport) writeErr(ctx context.Context, rw http.ResponseWriter, err error, req courierhttp.Request) {
	statusErr, ok := statuserror.IsStatusErr(err)
	if !ok {
		if errors.Is(err, context.Canceled) {
			// https://httpstatuses.com/499
			statusErr = statuserror.Wrap(err, 499, "ContextCanceled")
		} else {
			statusErr = statuserror.Wrap(err, http.StatusInternalServerError, "InternalServerError")
		}
	}

	statusErr = statusErr.AppendSource(req.ServiceName())

	if errResponseWriter := ErrResponseWriterFromContext(ctx); errResponseWriter != nil {
		errResponseWriter.WriteErr(ctx, rw, req, statusErr)
	} else {
		rw.Header().Set("Content-Type", "application/json; charset=utf-8")
		rw.WriteHeader(statusErr.StatusCode())
		_ = json.NewEncoder(rw).Encode(statusErr)
	}
}

func ErrResponseWriterFunc(fn func(ctx context.Context, rw http.ResponseWriter, req courierhttp.Request, statusErr *statuserror.StatusErr)) ErrResponseWriter {
	return &errResponseWriterFunc{fn: fn}
}

type errResponseWriterFunc struct {
	fn func(ctx context.Context, rw http.ResponseWriter, req courierhttp.Request, statusErr *statuserror.StatusErr)
}

func (e *errResponseWriterFunc) WriteErr(ctx context.Context, rw http.ResponseWriter, req courierhttp.Request, statusErr *statuserror.StatusErr) {
	e.fn(ctx, rw, req, statusErr)
}

type ErrResponseWriter interface {
	WriteErr(ctx context.Context, rw http.ResponseWriter, req courierhttp.Request, statusErr *statuserror.StatusErr)
}

type contextErrResponseWriter struct{}

func ContextWithErrResponseWriter(ctx context.Context, errResponseWriter ErrResponseWriter) context.Context {
	return context.WithValue(ctx, contextErrResponseWriter{}, errResponseWriter)
}

func ErrResponseWriterFromContext(ctx context.Context) ErrResponseWriter {
	if writeErrResp, ok := ctx.Value(contextErrResponseWriter{}).(ErrResponseWriter); ok {
		return writeErrResp
	}
	return nil
}

func (t *incomingTransport) UnmarshalOperator(ctx context.Context, info courierhttp.Request, op any) error {
	if err := t.decodeFromRequestInfo(ctx, info, op); err != nil {
		return err
	}
	return t.validate(op)
}

func (t *incomingTransport) validate(v interface{}) error {
	if canValidate, ok := v.(interface{ Validate() error }); ok {
		if err := canValidate.Validate(); err != nil {
			if est := err.(interface {
				ToFieldErrors() statuserror.ErrorFields
			}); ok {
				if errorFields := est.ToFieldErrors(); len(errorFields) > 0 {
					return (&badRequest{errorFields: errorFields}).Err()
				}
			}
			return err
		}
		return nil
	}

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	errSet := validator.NewErrorSet()

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			param := parameters[i]

			if param.Validator != nil {
				if err := param.Validator.Validate(param.FieldValue(rv)); err != nil {
					if param.In == "body" {
						errSet.AddErr(err, validator.Location(param.In))
					} else {
						errSet.AddErr(err, validator.Location(param.In), param.Name)
					}
				}
			}
		}
	}

	br := badRequestFromErrSet(errSet)

	if errSet.Err() == nil {
		return nil
	}

	return br.Err()
}

func (t *incomingTransport) decodeFromRequestInfo(ctx context.Context, info courierhttp.Request, v interface{}) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if rv.Kind() != reflect.Ptr {
		return errors.Errorf("decode target must be an ptr value")
	}

	rv = reflectx.Indirect(rv)

	if tpe := rv.Type(); tpe != t.Type {
		return errors.Errorf("unmatched request transformer, need %s but got %s", t.Type, tpe)
	}

	errSet := validator.NewErrorSet()

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			param := parameters[i]

			if param.In == "body" {
				if !param.TransformerOption.Strict || strings.HasPrefix(info.Header().Get("Content-Type"), param.Transformer.Names()[0]) {
					err := param.Transformer.DecodeFrom(ctx, info.Body(), param.FieldValue(rv).Addr(), textproto.MIMEHeader(info.Header()))
					if err != nil && err != io.EOF {
						errSet.AddErr(err, validator.Location(param.In))
					}
				}
				continue
			}

			var values []string

			if param.In == "meta" {
				// FIXME
			} else {
				values = info.Values(param.In, param.Name)
			}

			if len(values) > 0 {
				err := core.Wrap(param.Transformer, &param.TransformerOption.CommonOption).
					DecodeFrom(ctx, core.NewStringReaders(values), param.FieldValue(rv).Addr())

				if err != nil {
					errSet.AddErr(err, validator.Location(param.In), param.Name)
				}
			}
		}
	}

	if errSet.Err() == nil {
		return nil
	}

	return badRequestFromErrSet(errSet).Err()
}

func badRequestFromErrSet(set *validator.ErrorSet) *badRequest {
	br := &badRequest{}

	set.Flatten().Each(func(fieldErr *validator.FieldError) {
		if l, ok := fieldErr.Path[0].(validator.Location); ok {
			fe := &statuserror.ErrorField{
				In:    string(l),
				Field: fieldErr.Path[1:].String(),
				Msg:   fieldErr.Error.Error(),
			}
			br.errorFields = append(br.errorFields, fe)
		}
	})

	return br
}

type badRequest struct {
	errorFields statuserror.ErrorFields
	errTalk     bool
	msg         string
}

func (e *badRequest) EnableErrTalk() {
	e.errTalk = true
}

func (e *badRequest) SetMsg(msg string) {
	e.msg = msg
}

func (e *badRequest) AddErr(err error, nameOrIdx ...interface{}) {
	if len(nameOrIdx) > 1 {
		e.errorFields = append(e.errorFields, &statuserror.ErrorField{
			In:    nameOrIdx[0].(string),
			Field: validator.KeyPath(nameOrIdx[1:]).String(),
			Msg:   err.Error(),
		})
	}
}

func (e *badRequest) Err() error {
	if len(e.errorFields) == 0 {
		return nil
	}

	msg := e.msg
	if msg == "" {
		msg = "invalid parameters"
	}

	err := statuserror.
		Wrap(errors.New(""), http.StatusBadRequest, "badRequest").
		WithMsg(msg).
		AppendErrorFields(e.errorFields...)

	if e.errTalk {
		err = err.EnableErrTalk()
	}

	return err
}

type Upgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request) error
}

func (i *incomingTransport) WriteResponse(ctx context.Context, rw http.ResponseWriter, ret any, req courierhttp.Request) {
	if upgrader, ok := ret.(Upgrader); ok {
		if err := upgrader.Upgrade(rw, req.Underlying()); err != nil {
			i.writeErr(ctx, rw, err, req)
		}
		return
	}

	if err, ok := ret.(error); ok {
		i.writeErr(ctx, rw, err, req)
	} else {
		i.writeResp(ctx, rw, ret, req)
	}
}

func (i *incomingTransport) writeResp(ctx context.Context, rw http.ResponseWriter, ret any, req courierhttp.Request) {
	if err := courierhttp.Wrap(ret).(courierhttp.ResponseWriter).WriteResponse(req.Context(), rw, req); err != nil {
		i.writeErr(ctx, rw, err, req)
	}
}
