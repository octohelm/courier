package statuserr

import "net/http"

type NotFound struct{}

func (NotFound) StatusCode() int {
	return http.StatusNotFound
}
