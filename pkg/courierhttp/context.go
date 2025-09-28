package courierhttp

import (
	"net/http"
)

// +gengo:injectable:provider
type Request = http.Request

// +gengo:injectable:provider
type OperationInfo struct {
	Server

	ID     string
	Method string
	Route  string
}

// +gengo:injectable:provider
type OperationInfoProvider interface {
	GetOperation(id string) (OperationInfo, bool)
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
