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

	if p.Paths[path].Operations == nil {
		p.Paths[path].Operations = map[string]*OperationObject{}
	}

	p.Paths[path].Operations[strings.ToLower(method)] = op
}
