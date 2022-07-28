package statuserror

type StatusError interface {
	StatusErr() *StatusErr
	Error() string
}
