package statuserror

import (
	"bytes"
	"sort"
)

func NewErrorField(in string, field string, msg string) *ErrorField {
	return &ErrorField{
		In:    in,
		Field: field,
		Msg:   msg,
	}
}

type ErrorField struct {
	// field path
	// prop.slice[2].a
	Field string `json:"field" xml:"field"`
	// msg
	Msg string `json:"msg" xml:"msg"`
	// location
	// eq. body, query, header, path, formData
	In string `json:"in" xml:"in"`
}

func (s ErrorField) String() string {
	return s.Field + " in " + s.In + " - " + s.Msg
}

type ErrorFields []*ErrorField

func (fields ErrorFields) String() string {
	if len(fields) == 0 {
		return ""
	}

	sort.Sort(fields)

	buf := &bytes.Buffer{}
	buf.WriteString("<")
	for i, f := range fields {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(f.String())
	}
	buf.WriteString(">")
	return buf.String()
}

func (fields ErrorFields) Len() int {
	return len(fields)
}

func (fields ErrorFields) Swap(i, j int) {
	fields[i], fields[j] = fields[j], fields[i]
}

func (fields ErrorFields) Less(i, j int) bool {
	return fields[i].Field < fields[j].Field
}
