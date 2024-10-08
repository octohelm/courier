// copy from https://github.com/go-json-experiment/json/blob/master/fields.go
package jsonflags

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/go-json-experiment/json/jsontext"
)

const (
	nocase     = 1
	strictcase = 2
)

type fieldOptions struct {
	name       string
	quotedName string // quoted name per RFC 8785, section 3.2.2.2.
	hasName    bool
	casing     int8 // either 0, nocase, or strictcase
	inline     bool
	unknown    bool
	omitzero   bool
	omitempty  bool
	string     bool
	format     string
}

func parseFieldOptions(sf reflect.StructField) (out fieldOptions, ignored bool, err error) {
	tag, hasTag := sf.Tag.Lookup("json")
	if !hasTag {
		tag, hasTag = sf.Tag.Lookup("name")
	}

	// Check whether this field is explicitly ignored.
	if tag == "-" {
		return fieldOptions{}, true, nil
	}

	// Check whether this field is unexported.
	if !sf.IsExported() {
		// In contrast to v1, v2 no longer forwards exported fields from
		// embedded fields of unexported types since Go reflection does not
		// allow the same set of operations that are available in normal cases
		// of purely exported fields.
		// See https://go.dev/issue/21357 and https://go.dev/issue/24153.
		if sf.Anonymous {
			err = firstError(err, fmt.Errorf("embedded Go struct field %s of an unexported type must be explicitly ignored with a `json:\"-\"` tag", sf.Type.Name()))
		}
		// Tag options specified on an unexported field suggests user error.
		if hasTag {
			err = firstError(err, fmt.Errorf("unexported Go struct field %s cannot have non-ignored `json:%q` tag", sf.Name, tag))
		}
		return fieldOptions{}, true, err
	}

	// Determine the JSON member byName for this Go field. A user-specified byName
	// may be provided as either an identifier or a single-quoted string.
	// The single-quoted string allows arbitrary characters in the byName.
	// See https://go.dev/issue/2718 and https://go.dev/issue/3546.
	out.name = sf.Name // always starts with an uppercase character
	if len(tag) > 0 && !strings.HasPrefix(tag, ",") {
		// For better compatibility with v1, accept almost any unescaped byName.
		n := len(tag) - len(strings.TrimLeftFunc(tag, func(r rune) bool {
			return !strings.ContainsRune(",\\'\"`", r) // reserve comma, backslash, and quotes
		}))
		opt := tag[:n]
		if n == 0 {
			// Allow a single quoted string for arbitrary names.
			var err2 error
			opt, n, err2 = consumeTagOption(tag)
			if err2 != nil {
				err = firstError(err, fmt.Errorf("Go struct field %s has malformed `json` tag: %v", sf.Name, err2))
			}
		}
		out.hasName = true
		out.name = opt
		tag = tag[n:]
	}
	b, _ := jsontext.AppendQuote(nil, out.name)
	out.quotedName = string(b)

	// Handle any additional tag options (if any).
	var wasFormat bool
	seenOpts := make(map[string]bool)
	for len(tag) > 0 {
		// Consume comma delimiter.
		if tag[0] != ',' {
			err = firstError(err, fmt.Errorf("Go struct field %s has malformed `json` tag: invalid character %q before next option (expecting ',')", sf.Name, tag[0]))
		} else {
			tag = tag[len(","):]
			if len(tag) == 0 {
				err = firstError(err, fmt.Errorf("Go struct field %s has malformed `json` tag: invalid trailing ',' character", sf.Name))
				break
			}
		}

		// Consume and process the tag option.
		opt, n, err2 := consumeTagOption(tag)
		if err2 != nil {
			err = firstError(err, fmt.Errorf("Go struct field %s has malformed `json` tag: %v", sf.Name, err2))
		}
		rawOpt := tag[:n]
		tag = tag[n:]
		switch {
		case wasFormat:
			err = firstError(err, fmt.Errorf("Go struct field %s has `format` tag option that was not specified last", sf.Name))
		case strings.HasPrefix(rawOpt, "'") && strings.TrimFunc(opt, isLetterOrDigit) == "":
			err = firstError(err, fmt.Errorf("Go struct field %s has unnecessarily quoted appearance of `%s` tag option; specify `%s` instead", sf.Name, rawOpt, opt))
		}
		switch opt {
		case "nocase":
			out.casing |= nocase
		case "strictcase":
			out.casing |= strictcase
		case "inline":
			out.inline = true
		case "unknown":
			out.unknown = true
		case "omitzero":
			out.omitzero = true
		case "omitempty":
			out.omitempty = true
		case "string":
			out.string = true
		case "format":
			if !strings.HasPrefix(tag, ":") {
				err = firstError(err, fmt.Errorf("Go struct field %s is missing value for `format` tag option", sf.Name))
				break
			}
			tag = tag[len(":"):]
			opt, n, err2 := consumeTagOption(tag)
			if err2 != nil {
				err = firstError(err, fmt.Errorf("Go struct field %s has malformed value for `format` tag option: %v", sf.Name, err2))
				break
			}
			tag = tag[n:]
			out.format = opt
			wasFormat = true
		default:
			// Reject keys that resemble one of the supported options.
			// This catches invalid mutants such as "omitEmpty" or "omit_empty".
			normOpt := strings.ReplaceAll(strings.ToLower(opt), "_", "")
			switch normOpt {
			case "nocase", "strictcase", "inline", "unknown", "omitzero", "omitempty", "string", "format":
				err = firstError(err, fmt.Errorf("Go struct field %s has invalid appearance of `%s` tag option; specify `%s` instead", sf.Name, opt, normOpt))
			}

			// NOTE: Everything else is ignored. This does not mean it is
			// forward compatible to insert arbitrary tag options since
			// a future version of this package may understand that tag.
		}

		// Reject duplicates.
		switch {
		case out.casing == nocase|strictcase:
			err = firstError(err, fmt.Errorf("Go struct field %s cannot have both `nocase` and `structcase` tag options", sf.Name))
		case seenOpts[opt]:
			err = firstError(err, fmt.Errorf("Go struct field %s has duplicate appearance of `%s` tag option", sf.Name, rawOpt))
		}
		seenOpts[opt] = true
	}
	return out, false, err
}

func consumeTagOption(in string) (string, int, error) {
	// For legacy compatibility with v1, assume options are comma-separated.
	i := strings.IndexByte(in, ',')
	if i < 0 {
		i = len(in)
	}

	switch r, _ := utf8.DecodeRuneInString(in); {
	// Option as a Go identifier.
	case r == '_' || unicode.IsLetter(r):
		n := len(in) - len(strings.TrimLeftFunc(in, isLetterOrDigit))
		return in[:n], n, nil
	// Option as a single-quoted string.
	case r == '\'':
		// The grammar is nearly identical to a double-quoted Go string literal,
		// but uses single quotes as the terminators. The reason for a custom
		// grammar is because both backtick and double quotes cannot be used
		// verbatim in a struct tag.
		//
		// Convert a single-quoted string to a double-quote string and rely on
		// strconv.Unquote to handle the rest.
		var inEscape bool
		b := []byte{'"'}
		n := len(`'`)
		for len(in) > n {
			r, rn := utf8.DecodeRuneInString(in[n:])
			switch {
			case inEscape:
				if r == '\'' {
					b = b[:len(b)-1] // remove escape character: `\'` => `'`
				}
				inEscape = false
			case r == '\\':
				inEscape = true
			case r == '"':
				b = append(b, '\\') // insert escape character: `"` => `\"`
			case r == '\'':
				b = append(b, '"')
				n += len(`'`)
				out, err := strconv.Unquote(string(b))
				if err != nil {
					return in[:i], i, fmt.Errorf("invalid single-quoted string: %s", in[:n])
				}
				return out, n, nil
			}
			b = append(b, in[n:][:rn]...)
			n += rn
		}
		if n > 10 {
			n = 10 // limit the amount of context printed in the error
		}
		return in[:i], i, fmt.Errorf("single-quoted string not terminated: %s...", in[:n])
	case len(in) == 0:
		return in[:i], i, io.ErrUnexpectedEOF
	default:
		return in[:i], i, fmt.Errorf("invalid character %q at start of option (expecting Unicode letter or single quote)", r)
	}
}

func isLetterOrDigit(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r)
}
