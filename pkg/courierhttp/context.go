package courierhttp

import (
	"net/http"
)

// +gengo:injectable:provider
type HttpRequest struct {
	*http.Request
}

// +gengo:injectable:provider
type OperationInfo struct {
	Server
	ID     string
	Method string
	Route  string
}

func (s OperationInfo) UserAgent() string {
	id := s.ID
	if id == "" {
		id = "Unknown"
	}
	return s.Server.UserAgent() + " (" + id + ")"
}

type Server struct {
	Name    string
	Version string
}

func (s Server) UserAgent() string {
	if s.Version == "" {
		return s.Name
	}
	return s.Name + "/" + s.Version
}
