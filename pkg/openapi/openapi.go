package openapi

import (
	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
	"strings"
)

type Payload struct {
	OpenAPI
}

func (p Payload) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.OpenAPI)
}

func (p *Payload) UnmarshalJSON(data []byte) (err error) {
	openapi := &OpenAPI{}
	if err := jsonschema.Unmarshal(data, openapi); err != nil {
		return err
	}
	*p = Payload{
		OpenAPI: *openapi,
	}
	return nil
}

func NewOpenAPI() *OpenAPI {
	openAPI := &OpenAPI{}
	openAPI.OpenAPI = "3.1.0"
	return openAPI
}

type OpenAPI struct {
	OpenAPI string `json:"openapi"`

	InfoObject       `json:"info"`
	ComponentsObject `json:"components"`

	Paths map[string]*PathItemObject `json:"paths"`

	jsonschema.Ext
}

func (p *OpenAPI) AddOperation(method string, path string, op *OperationObject) {
	if p.Paths == nil {
		p.Paths = map[string]*PathItemObject{}
	}

	if p.Paths[path] == nil {
		p.Paths[path] = &PathItemObject{}
	}

	switch strings.ToLower(method) {
	case "get":
		p.Paths[path].GET = op
	case "post":
		p.Paths[path].POST = op
	case "put":
		p.Paths[path].PUT = op
	case "patch":
		p.Paths[path].PATCH = op
	case "delete":
		p.Paths[path].DELETE = op
	case "trace":
		p.Paths[path].TRACE = op
	case "head":
		p.Paths[path].HEAD = op
	case "options":
		p.Paths[path].OPTIONS = op
	}
}
