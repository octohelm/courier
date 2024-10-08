package transport

import (
	"context"
	"io"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"sync"

	"github.com/go-courier/logr"

	"github.com/octohelm/courier/pkg/courierhttp"
	"github.com/octohelm/courier/pkg/statuserror"
	"github.com/octohelm/courier/pkg/transformer"
	"github.com/octohelm/courier/pkg/transformer/core"
	reflectx "github.com/octohelm/x/reflect"
	typex "github.com/octohelm/x/types"
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

func (t *incomingTransport) UnmarshalOperator(ctx context.Context, info courierhttp.Request, op any) error {
	if err := t.decodeFromRequestInfo(ctx, info, op); err != nil {
		return err
	}
	return t.validate(op)
}

func (t *incomingTransport) validate(v any) error {
	if canValidate, ok := v.(interface{ Validate() error }); ok {
		if err := canValidate.Validate(); err != nil {
			return err
		}
		return nil
	}

	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	var finalError error

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			param := parameters[i]

			if param.Validator != nil {
				if err := param.Validator.Validate(param.FieldValue(rv)); err != nil {
					if param.In == "body" {
						finalError = statuserror.Append(finalError, statuserror.ParameterError(param.In))
					} else {
						finalError = statuserror.Append(finalError, statuserror.ParameterError(param.In, param.Name))
					}
				}
			}
		}
	}

	return finalError
}

func (t *incomingTransport) decodeFromRequestInfo(ctx context.Context, info courierhttp.Request, v any) error {
	rv, ok := v.(reflect.Value)
	if !ok {
		rv = reflect.ValueOf(v)
	}

	if rv.Kind() != reflect.Ptr {
		return fmt.Errorf("decode target must be an ptr value")
	}

	rv = reflectx.Indirect(rv)

	if tpe := rv.Type(); tpe != t.Type {
		return fmt.Errorf("unmatched request transformer, need %s but got %s", t.Type, tpe)
	}

	var finalError error

	for in := range t.InParameters {
		parameters := t.InParameters[in]

		for i := range parameters {
			param := parameters[i]

			if param.In == "body" {
				if !param.TransformerOption.Strict || strings.HasPrefix(info.Header().Get("Content-Type"), param.Transformer.Names()[0]) {
					err := param.Transformer.DecodeFrom(ctx, info.Body(), param.FieldValue(rv).Addr(), textproto.MIMEHeader(info.Header()))
					if err != nil && err != io.EOF {
						finalError = statuserror.Append(finalError, statuserror.ParameterError(param.In))
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
				paramValue := param.FieldValue(rv)
				if paramValue.Kind() != reflect.Ptr {
					paramValue = paramValue.Addr()
				}

				err := core.Wrap(param.Transformer, &param.TransformerOption.CommonOption).
					DecodeFrom(ctx, core.NewStringReaders(values), paramValue)

				if err != nil {
					finalError = statuserror.Append(finalError, statuserror.ParameterError(param.In, param.Name))
				}
			}
		}
	}

	return finalError
}

//func badRequestFromErrSet(set *validator.ErrorSet) *badRequest {
//	br := &badRequest{}
//
//	set.Flatten().Each(func(fieldErr *validator.FieldError) {
//		if l, ok := fieldErr.Path[0].(validator.Location); ok {
//			fe := &statuserror.ErrorField{
//				In:    string(l),
//				Field: fieldErr.Path[1:].String(),
//				Msg:   fieldErr.Error.Error(),
//			}
//			br.errorFields = append(br.errorFields, fe)
//		}
//	})
//
//	return br
//}

//type badRequest struct {
//	errorFields statuserror.ErrorFields
//	msg         string
//}
//
//func (e *badRequest) SetMsg(msg string) {
//	e.msg = msg
//}
//
//func (e *badRequest) AddErr(err error, nameOrIdx ...any) {
//	if len(nameOrIdx) > 1 {
//		e.errorFields = append(e.errorFields, &statuserror.ErrorField{
//			In:    nameOrIdx[0].(string),
//			Field: validator.KeyPath(nameOrIdx[1:]).String(),
//			Msg:   err.Error(),
//		})
//	}
//}
//
//func (e *badRequest) Err() error {
//	if len(e.errorFields) == 0 {
//		return nil
//	}
//
//	msg := e.msg
//	if msg == "" {
//		msg = "invalid parameters"
//	}
//
//	err := statuserror.
//		Wrap(errors.New(""), http.StatusBadRequest, "badRequest").
//		WithMsg(msg).
//		AppendErrorFields(e.errorFields...)
//
//
//	return err
//}

type Upgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request) error
}

func (i *incomingTransport) WriteResponse(ctx context.Context, rw http.ResponseWriter, ret any, req courierhttp.Request) {
	if upgrader, ok := ret.(Upgrader); ok {
		if err := upgrader.Upgrade(rw, req.Underlying()); err != nil {
			i.writeErrResp(ctx, rw, err, req)
		}
		return
	}

	if err, ok := ret.(error); ok {
		i.writeErrResp(ctx, rw, err, req)
	} else {
		i.writeResp(ctx, rw, ret, req)
	}
}

func (i *incomingTransport) writeResp(ctx context.Context, rw http.ResponseWriter, ret any, req courierhttp.Request) {
	if err := courierhttp.Wrap(ret).(courierhttp.ResponseWriter).WriteResponse(ctx, rw, req); err != nil {
		logr.FromContext(ctx).Error(err)
	}
}

func (i *incomingTransport) writeErrResp(ctx context.Context, rw http.ResponseWriter, err error, req courierhttp.Request) {
	if err := courierhttp.WrapError(err).(courierhttp.ResponseWriter).WriteResponse(ctx, rw, req); err != nil {
		logr.FromContext(ctx).Error(err)
	}
}
