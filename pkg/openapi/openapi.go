package openapi

import (
	"strings"

	"github.com/go-json-experiment/json"
	"github.com/octohelm/courier/pkg/openapi/internal"
	"github.com/octohelm/courier/pkg/openapi/jsonschema"
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

	Paths internal.Record[string, *PathItemObject] `json:"paths"`

	jsonschema.Ext
}

func (p *OpenAPI) AddOperation(method string, path string, op *OperationObject) {
	operations, ok := p.Paths.Get(path)
	if !ok {
		operations = &PathItemObject{}
		p.Paths.Set(path, operations)
	}

	operations.Set(strings.ToLower(method), op)
}
