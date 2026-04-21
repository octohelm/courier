package statuserror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-json-experiment/json/jsontext"
	. "github.com/octohelm/x/testing/v2"
)

type locatedError struct{}

func (locatedError) Error() string    { return "bad input" }
func (locatedError) StatusCode() int  { return http.StatusBadRequest }
func (locatedError) Location() string { return "query.limit" }

type pointerError struct{}

func (pointerError) Error() string                 { return "invalid body" }
func (pointerError) JSONPointer() jsontext.Pointer { return "/items/0" }

type exportedError struct{}

func (exportedError) Error() string { return "exported" }

func TestWrapAndErrCode(t0 *testing.T) {
	Then(t0, "包装错误与错误码推导符合预期",
		Expect(Wrap(nil, http.StatusBadRequest, "BAD"), Equal[error](nil)),
		ExpectMust(func() error {
			err := Wrap(errors.New("boom"), http.StatusConflict, "CONFLICT")
			if err.Error() != `CONFLICT{message="boom",statusCode=409}` {
				return fmt.Errorf("unexpected wrapped error %s", err.Error())
			}
			if err.(WithStatusCode).StatusCode() != http.StatusConflict {
				return fmt.Errorf("unexpected status %d", err.(WithStatusCode).StatusCode())
			}
			if err.(WithErrCode).ErrCode() != "CONFLICT" {
				return fmt.Errorf("unexpected code %s", err.(WithErrCode).ErrCode())
			}
			if errors.Unwrap(err).Error() != "boom" {
				return fmt.Errorf("unexpected unwrap %v", errors.Unwrap(err))
			}
			return nil
		}),
		Expect(ErrCodeFor[exportedError](), Equal("")),
		Expect(ErrCodeOf(BadRequest{}), Equal("statuserror.BadRequest")),
		Expect(ErrCodeOf(struct{}{}), Equal("")),
	)
}

func TestAllAndAsErrorResponse(t0 *testing.T) {
	joined := errors.Join(
		locatedError{},
		pointerError{},
		Wrap(errors.New("failed"), http.StatusConflict, "CONFLICT"),
	)

	Then(t0, "错误链遍历与错误响应转换符合预期",
		ExpectMust(func() error {
			all := make([]error, 0)
			for err := range All(joined) {
				all = append(all, err)
			}
			if len(all) != 3 {
				return fmt.Errorf("unexpected error count %d", len(all))
			}
			return nil
		}),
		ExpectMust(func() error {
			resp := AsErrorResponse(joined, "courier")
			if resp == nil {
				return errors.New("nil response")
			}
			if resp.Code != http.StatusBadRequest {
				return fmt.Errorf("unexpected code %d", resp.Code)
			}
			if resp.Msg != http.StatusText(http.StatusBadRequest) {
				return fmt.Errorf("unexpected msg %q", resp.Msg)
			}
			if len(resp.Errors) != 2 {
				return fmt.Errorf("unexpected errors %d", len(resp.Errors))
			}
			if resp.Errors[0].Location != "query.limit" {
				return fmt.Errorf("unexpected location %q", resp.Errors[0].Location)
			}
			if resp.Errors[0].Pointer != "/items/0" {
				return fmt.Errorf("unexpected pointer %q", resp.Errors[0].Pointer)
			}
			if resp.Errors[1].Code != "CONFLICT" {
				return fmt.Errorf("unexpected code %q", resp.Errors[1].Code)
			}
			return nil
		}),
		ExpectMust(func() error {
			resp := AsErrorResponse(errors.New("plain"), "courier")
			if resp.Code != http.StatusInternalServerError || resp.Msg != "plain" {
				return fmt.Errorf("unexpected plain response %#v", resp)
			}
			return nil
		}),
	)
}

func TestDescriptorAndErrorResponseHelpers(t0 *testing.T) {
	raw := []byte(`{"code":400001,"msg":"bad request","errors":[{"code":"INVALID_PARAMETER","message":"missing","location":"query"}],"title":"ignored","detail":"also ignored"}`)

	Then(t0, "描述对象与错误响应辅助方法符合预期",
		ExpectMust(func() error {
			var d Descriptor
			if err := d.UnmarshalErrorResponse(http.StatusBadRequest, raw); err != nil {
				return err
			}
			if d.Code != "INVALID_PARAMETER" || d.Message != "also ignored" {
				return fmt.Errorf("unexpected descriptor %#v", d)
			}
			if d.StatusCode() != http.StatusBadRequest {
				return fmt.Errorf("unexpected status %d", d.StatusCode())
			}
			return nil
		}),
		ExpectMust(func() error {
			var d Descriptor
			if err := d.UnmarshalErrorResponse(http.StatusBadGateway, []byte("gateway timeout")); err != nil {
				return err
			}
			if d.Message != "gateway timeout" {
				return fmt.Errorf("unexpected message %q", d.Message)
			}
			return nil
		}),
		Expect((&ErrorResponse{Code: 400001}).StatusCode(), Equal(http.StatusBadRequest)),
		Expect((&ErrorResponse{Code: http.StatusBadRequest}).StatusCode(), Equal(http.StatusBadRequest)),
		ExpectMust(func() error {
			resp := &ErrorResponse{
				Errors: []*Descriptor{{Code: "A"}, {Code: "B"}},
			}
			errs := resp.Unwrap()
			if len(errs) != 2 {
				return fmt.Errorf("unexpected unwrap len %d", len(errs))
			}
			return nil
		}),
		Expect((&Descriptor{Code: "INVALID_PARAMETER", Message: "missing"}).Error(), Equal(`INVALID_PARAMETER{message="missing"}`)),
		Expect(wrapError{err: errors.New("boom")}.Error(), Equal("boom")),
	)
}

func TestHTTPStatusErrors(t0 *testing.T) {
	cases := []struct {
		name string
		err  WithStatusCode
		code int
	}{
		{"ClientClosedRequest", ClientClosedRequest{}, 499},
		{"BadRequest", BadRequest{}, http.StatusBadRequest},
		{"Unauthorized", Unauthorized{}, http.StatusUnauthorized},
		{"PaymentRequired", PaymentRequired{}, http.StatusPaymentRequired},
		{"Forbidden", Forbidden{}, http.StatusForbidden},
		{"NotFound", NotFound{}, http.StatusNotFound},
		{"MethodNotAllowed", MethodNotAllowed{}, http.StatusMethodNotAllowed},
		{"NotAcceptable", NotAcceptable{}, http.StatusNotAcceptable},
		{"ProxyAuthRequired", ProxyAuthRequired{}, http.StatusProxyAuthRequired},
		{"RequestTimeout", RequestTimeout{}, http.StatusRequestTimeout},
		{"Conflict", Conflict{}, http.StatusConflict},
		{"Gone", Gone{}, http.StatusGone},
		{"LengthRequired", LengthRequired{}, http.StatusLengthRequired},
		{"PreconditionFailed", PreconditionFailed{}, http.StatusPreconditionFailed},
		{"RequestEntityTooLarge", RequestEntityTooLarge{}, http.StatusRequestEntityTooLarge},
		{"RequestURITooLong", RequestURITooLong{}, http.StatusRequestURITooLong},
		{"UnsupportedMediaType", UnsupportedMediaType{}, http.StatusUnsupportedMediaType},
		{"RequestedRangeNotSatisfiable", RequestedRangeNotSatisfiable{}, http.StatusRequestedRangeNotSatisfiable},
		{"ExpectationFailed", ExpectationFailed{}, http.StatusExpectationFailed},
		{"Teapot", Teapot{}, http.StatusTeapot},
		{"MisdirectedRequest", MisdirectedRequest{}, http.StatusMisdirectedRequest},
		{"UnprocessableEntity", UnprocessableEntity{}, http.StatusUnprocessableEntity},
		{"Locked", Locked{}, http.StatusLocked},
		{"FailedDependency", FailedDependency{}, http.StatusFailedDependency},
		{"TooEarly", TooEarly{}, http.StatusTooEarly},
		{"UpgradeRequired", UpgradeRequired{}, http.StatusUpgradeRequired},
		{"PreconditionRequired", PreconditionRequired{}, http.StatusPreconditionRequired},
		{"TooManyRequests", TooManyRequests{}, http.StatusTooManyRequests},
		{"RequestHeaderFieldsTooLarge", RequestHeaderFieldsTooLarge{}, http.StatusRequestHeaderFieldsTooLarge},
		{"UnavailableForLegalReasons", UnavailableForLegalReasons{}, http.StatusUnavailableForLegalReasons},
		{"InternalServerError", InternalServerError{}, http.StatusInternalServerError},
		{"NotImplemented", NotImplemented{}, http.StatusNotImplemented},
		{"BadGateway", BadGateway{}, http.StatusBadGateway},
		{"ServiceUnavailable", ServiceUnavailable{}, http.StatusServiceUnavailable},
		{"GatewayTimeout", GatewayTimeout{}, http.StatusGatewayTimeout},
		{"HTTPVersionNotSupported", HTTPVersionNotSupported{}, http.StatusHTTPVersionNotSupported},
		{"VariantAlsoNegotiates", VariantAlsoNegotiates{}, http.StatusVariantAlsoNegotiates},
		{"InsufficientStorage", InsufficientStorage{}, http.StatusInsufficientStorage},
		{"LoopDetected", LoopDetected{}, http.StatusLoopDetected},
		{"NotExtended", NotExtended{}, http.StatusNotExtended},
		{"NetworkAuthenticationRequired", NetworkAuthenticationRequired{}, http.StatusNetworkAuthenticationRequired},
	}

	checkers := make([]Checker, 0, len(cases))
	for _, c := range cases {
		checkers = append(checkers, Expect(c.err.StatusCode(), Equal(c.code)))
	}

	Then(t0, "预定义 HTTP 状态错误映射正确", checkers...)
}

func TestRuntimeDocAndSequenceBranches(t0 *testing.T) {
	Then(t0, "运行时文档与错误链其他分支可覆盖",
		ExpectMust(func() error {
			if doc, ok := (&Descriptor{}).RuntimeDoc("Code"); !ok || len(doc) == 0 {
				return fmt.Errorf("missing descriptor doc")
			}
			for _, name := range []string{"Message", "Description", "Location", "Pointer", "Source", "Errors", "Extra", "Status"} {
				if _, ok := (&Descriptor{}).RuntimeDoc(name); !ok {
					return fmt.Errorf("missing descriptor doc for %s", name)
				}
			}
			if doc, ok := (&Descriptor{}).RuntimeDoc(); !ok || len(doc) != 0 {
				return fmt.Errorf("unexpected descriptor root doc %v", doc)
			}
			if _, ok := (&Descriptor{}).RuntimeDoc("Unknown"); ok {
				return fmt.Errorf("unexpected descriptor doc hit")
			}
			if doc, ok := (&ErrorResponse{}).RuntimeDoc("Msg"); !ok || len(doc) == 0 {
				return fmt.Errorf("missing error response doc")
			}
			for _, name := range []string{"Code", "Errors"} {
				if _, ok := (&ErrorResponse{}).RuntimeDoc(name); !ok {
					return fmt.Errorf("missing error response doc for %s", name)
				}
			}
			if doc, ok := (&ErrorResponse{}).RuntimeDoc(); !ok || len(doc) != 0 {
				return fmt.Errorf("unexpected error response root doc %v", doc)
			}
			if _, ok := (&ErrorResponse{}).RuntimeDoc("Unknown"); ok {
				return fmt.Errorf("unexpected error response doc hit")
			}
			if doc, ok := new(IntOrString).RuntimeDoc(); !ok || len(doc) != 0 {
				return fmt.Errorf("unexpected int or string doc %v", doc)
			}
			if doc, ok := runtimeDoc(&Descriptor{}, "前缀:", "Code"); !ok || len(doc) == 0 || doc[0] != "前缀:错误编码" {
				return fmt.Errorf("unexpected prefixed doc %v", doc)
			}
			if _, ok := runtimeDoc(struct{}{}, "", "Code"); ok {
				return fmt.Errorf("unexpected runtime doc hit")
			}
			return nil
		}),
		ExpectMust(func() error {
			count := 0
			for range All(errors.Join(nil, errors.New("x"))) {
				count++
			}
			if count != 1 {
				return fmt.Errorf("unexpected join count %d", count)
			}
			count = 0
			for range All(nil) {
				count++
			}
			if count != 0 {
				return fmt.Errorf("unexpected nil count %d", count)
			}
			return nil
		}),
		Expect(isVersionSegment("v2"), Equal(true)),
		Expect(isVersionSegment("vx"), Equal(false)),
	)
}
