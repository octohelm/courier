package transport

import "net/http"

type Upgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request) error
}
