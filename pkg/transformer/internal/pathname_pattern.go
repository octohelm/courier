package internal

import (
	"strings"

	"github.com/julienschmidt/httprouter"
	pkgerrors "github.com/pkg/errors"
)

func ParamsFromMap(m map[string]string) httprouter.Params {
	params := httprouter.Params{}
	for k, v := range m {
		params = append(params, httprouter.Param{
			Key:   k,
			Value: v,
		})
	}
	return params
}

func NewPathnamePattern(p string) *pathnamePattern {
	parts := toPathParts(p)

	idxKeys := map[int]string{}

	for i, p := range parts {
		if p[0] == ':' {
			idxKeys[i] = p[1:]
		}
	}

	return &pathnamePattern{
		parts,
		idxKeys,
	}
}

type pathnamePattern struct {
	parts   []string
	idxKeys map[int]string
}

func (pattern *pathnamePattern) String() string {
	return "/" + strings.Join(pattern.parts, "/")
}

func (pattern *pathnamePattern) Stringify(params map[string]string) string {
	if len(pattern.idxKeys) == 0 {
		return pattern.String()
	}

	parts := append([]string{}, pattern.parts...)

	for idx, key := range pattern.idxKeys {
		v := params[key]
		if v == "" {
			v = "-"
		}
		parts[idx] = v
	}

	return (&pathnamePattern{parts: parts}).String()
}

func (pattern *pathnamePattern) Parse(pathname string) (params map[string]string, err error) {
	parts := toPathParts(pathname)

	if len(parts) != len(pattern.parts) {
		return nil, pkgerrors.Errorf("pathname %s is not match %s", pathname, pattern)
	}

	params = map[string]string{}

	for idx, part := range pattern.parts {
		if key, ok := pattern.idxKeys[idx]; ok {
			params[key] = parts[idx]
		} else if part != parts[idx] {
			return nil, pkgerrors.Errorf("pathname %s is not match %s", pathname, pattern)
		}
	}

	return
}

func toPathParts(p string) []string {
	p = httprouter.CleanPath(p)
	if p[0] == '/' {
		p = p[1:]
	}
	if p == "" {
		return make([]string, 0)
	}
	return strings.Split(p, "/")
}
