package jsonschema

import (
	"encoding/json"
	"strings"
)

type VendorExtensible struct {
	Extensions map[string]any
}

func (v *VendorExtensible) AddExtension(key string, value any) {
	if value == nil {
		return
	}
	if v.Extensions == nil {
		v.Extensions = make(map[string]any)
	}
	v.Extensions[key] = value
}

func (v VendorExtensible) MarshalJSON() ([]byte, error) {
	values := make(map[string]any)
	for k := range v.Extensions {
		if strings.HasPrefix(strings.ToLower(k), "x-") {
			values[k] = v.Extensions[k]
		}
	}
	return json.Marshal(values)
}

func (v *VendorExtensible) UnmarshalJSON(data []byte) error {
	var d map[string]any
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	for k := range d {
		if strings.HasPrefix(strings.ToLower(k), "x-") {
			if v.Extensions == nil {
				v.Extensions = map[string]any{}
			}
			v.Extensions[k] = d[k]
		}
	}
	return nil
}
