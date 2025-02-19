package jsonschema

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// openapi:strfmt uri
type URIString url.URL

func (u *URIString) UnmarshalText(b []byte) error {
	x, err := url.Parse(string(b))
	if err != nil {
		return err
	}
	*u = URIString(*x)
	return nil
}

func (u URIString) MarshalText() ([]byte, error) {
	x := url.URL(u)
	return []byte(x.String()), nil
}

func ParseURIReferenceString(u string) (*URIReferenceString, error) {
	x, err := url.Parse(u)
	if err != nil {
		return nil, err
	}
	return (*URIReferenceString)(x), nil
}

// openapi:strfmt uri-reference
type URIReferenceString url.URL

func (u *URIReferenceString) RefName() string {
	if u.Fragment == "" {
		return ""
	}
	// last part
	parts := strings.Split(u.Fragment, "/")
	return parts[len(parts)-1]
}

func (u *URIReferenceString) UnmarshalText(b []byte) error {
	x, err := url.Parse(string(b))
	if err != nil {
		return err
	}
	*u = URIReferenceString(*x)
	return nil
}

func (u URIReferenceString) MarshalText() ([]byte, error) {
	x := url.URL(u)

	if x.Scheme == "" {
		var buf strings.Builder

		if x.Path != "" {
			buf.WriteString(x.Path)
		}

		if x.RawQuery != "" {
			buf.WriteString("?")
			buf.WriteString(x.RawQuery)
		}

		if x.Fragment != "" {
			buf.WriteString("#")
			buf.WriteString(x.Fragment)
		}

		return []byte(buf.String()), nil
	}

	return []byte(x.String()), nil
}

// openapi:strfmt anchor
type AnchorString string

var (
	anchorPattern   = "^[A-Za-z_][-A-Za-z0-9._]*$"
	anchorPatternRe = regexp.MustCompile(anchorPattern)
)

func (s *AnchorString) UmarshalText(text []byte) error {
	if anchorPatternRe.Match(text) {
		*s = AnchorString(text)
	}
	return errors.New("invalid anchor string")
}

func (AnchorString) OpenAPISchema() Schema {
	return &StringType{
		Type:    "string",
		Pattern: anchorPattern,
	}
}
