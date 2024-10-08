package statuserror

import "net/http"

type ClientClosedRequest struct{}

func (ClientClosedRequest) StatusCode() int {
	return 499
}

type BadRequest struct{}

func (BadRequest) StatusCode() int {
	return http.StatusBadRequest
}

type Unauthorized struct{}

func (Unauthorized) StatusCode() int {
	return http.StatusUnauthorized
}

type PaymentRequired struct{}

func (PaymentRequired) StatusCode() int {
	return http.StatusPaymentRequired
}

type Forbidden struct{}

func (Forbidden) StatusCode() int {
	return http.StatusForbidden
}

type NotFound struct{}

func (NotFound) StatusCode() int {
	return http.StatusNotFound
}

type MethodNotAllowed struct{}

func (MethodNotAllowed) StatusCode() int {
	return http.StatusMethodNotAllowed
}

type NotAcceptable struct{}

func (NotAcceptable) StatusCode() int {
	return http.StatusNotAcceptable
}

type ProxyAuthRequired struct{}

func (ProxyAuthRequired) StatusCode() int {
	return http.StatusProxyAuthRequired
}

type RequestTimeout struct{}

func (RequestTimeout) StatusCode() int {
	return http.StatusRequestTimeout
}

type Conflict struct{}

func (Conflict) StatusCode() int {
	return http.StatusConflict
}

type Gone struct{}

func (Gone) StatusCode() int {
	return http.StatusGone
}

type LengthRequired struct{}

func (LengthRequired) StatusCode() int {
	return http.StatusLengthRequired
}

type PreconditionFailed struct{}

func (PreconditionFailed) StatusCode() int {
	return http.StatusPreconditionFailed
}

type RequestEntityTooLarge struct{}

func (RequestEntityTooLarge) StatusCode() int {
	return http.StatusRequestEntityTooLarge
}

type RequestURITooLong struct{}

func (RequestURITooLong) StatusCode() int {
	return http.StatusRequestURITooLong
}

type UnsupportedMediaType struct{}

func (UnsupportedMediaType) StatusCode() int {
	return http.StatusUnsupportedMediaType
}

type RequestedRangeNotSatisfiable struct{}

func (RequestedRangeNotSatisfiable) StatusCode() int {
	return http.StatusRequestedRangeNotSatisfiable
}

type ExpectationFailed struct{}

func (ExpectationFailed) StatusCode() int {
	return http.StatusExpectationFailed
}

type Teapot struct{}

func (Teapot) StatusCode() int {
	return http.StatusTeapot
}

type MisdirectedRequest struct{}

func (MisdirectedRequest) StatusCode() int {
	return http.StatusMisdirectedRequest
}

type UnprocessableEntity struct{}

func (UnprocessableEntity) StatusCode() int {
	return http.StatusUnprocessableEntity
}

type Locked struct{}

func (Locked) StatusCode() int {
	return http.StatusLocked
}

type FailedDependency struct{}

func (FailedDependency) StatusCode() int {
	return http.StatusFailedDependency
}

type TooEarly struct{}

func (TooEarly) StatusCode() int {
	return http.StatusTooEarly
}

type UpgradeRequired struct{}

func (UpgradeRequired) StatusCode() int {
	return http.StatusUpgradeRequired
}

type PreconditionRequired struct{}

func (PreconditionRequired) StatusCode() int {
	return http.StatusPreconditionRequired
}

type TooManyRequests struct{}

func (TooManyRequests) StatusCode() int {
	return http.StatusTooManyRequests
}

type RequestHeaderFieldsTooLarge struct{}

func (RequestHeaderFieldsTooLarge) StatusCode() int {
	return http.StatusRequestHeaderFieldsTooLarge
}

type UnavailableForLegalReasons struct{}

func (UnavailableForLegalReasons) StatusCode() int {
	return http.StatusUnavailableForLegalReasons
}

type InternalServerError struct{}

func (InternalServerError) StatusCode() int {
	return http.StatusInternalServerError
}

type NotImplemented struct{}

func (NotImplemented) StatusCode() int {
	return http.StatusNotImplemented
}

type BadGateway struct{}

func (BadGateway) StatusCode() int {
	return http.StatusBadGateway
}

type ServiceUnavailable struct{}

func (ServiceUnavailable) StatusCode() int {
	return http.StatusServiceUnavailable
}

type GatewayTimeout struct{}

func (GatewayTimeout) StatusCode() int {
	return http.StatusGatewayTimeout
}

type HTTPVersionNotSupported struct{}

func (HTTPVersionNotSupported) StatusCode() int {
	return http.StatusHTTPVersionNotSupported
}

type VariantAlsoNegotiates struct{}

func (VariantAlsoNegotiates) StatusCode() int {
	return http.StatusVariantAlsoNegotiates
}

type InsufficientStorage struct{}

func (InsufficientStorage) StatusCode() int {
	return http.StatusInsufficientStorage
}

type LoopDetected struct{}

func (LoopDetected) StatusCode() int {
	return http.StatusLoopDetected
}

type NotExtended struct{}

func (NotExtended) StatusCode() int {
	return http.StatusNotExtended
}

type NetworkAuthenticationRequired struct{}

func (NetworkAuthenticationRequired) StatusCode() int {
	return http.StatusNetworkAuthenticationRequired
}
