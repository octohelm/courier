package pathpattern

import (
	"fmt"
	"iter"
	"path"
	"strings"
)

func NormalizePath(p string) string {
	parts := splitPath(path.Clean(p))

	processed := make([]string, len(parts))

	for i := range processed {
		part := parts[i]

		// julienschmidt/httprouter style
		if strings.HasPrefix(part, ":") {
			processed[i] = fmt.Sprintf("{%s}", part[1:])
			continue
		}
		if strings.HasPrefix(part, "*") {
			processed[i] = fmt.Sprintf("{%s...}", part[1:])
			continue
		}

		processed[i] = part
	}

	return "/" + strings.Join(processed, "/")
}

type PathEncoder interface {
	Encode(map[string]string) string
}

type PathValuesGetter interface {
	PathValues(p string) (Values, error)
}

type Pattern interface {
	PathEncoder
	PathValuesGetter
}

func Parse(p string) Segments {
	parts := splitPath(p)

	ss := make(Segments, len(parts))

	for i, part := range parts {
		if len(part) > 0 && part[0] == '{' {

			named := namedSegment{
				name: part[1 : len(part)-1],
			}

			if strings.HasSuffix(named.name, "...") {
				named.name = named.name[0 : len(named.name)-3]
				named.multiple = true
			}

			ss[i] = named
			continue
		}

		ss[i] = segment(part)
	}

	return ss
}

type ValueSetter interface {
	Set(key string, value string)
}

type Values map[string]string

func (values Values) Set(key string, value string) {
	values[key] = value
}

func (ss Segments) PathValues(pathname string) (Values, error) {
	params := Values{}

	_, ok := ss.MatchTo(params, pathname)
	if !ok {
		return nil, fmt.Errorf("pathname %s is not match %s", pathname, ss)
	}

	return params, nil
}

func (ss Segments) MatchTo(setter ValueSetter, pathname string) (string, bool) {
	return createMatcher(ss, "").MatchTo(setter, pathname)
}

type Segments []Segment

func (ss Segments) Chunk() iter.Seq[Segments] {
	return func(yield func(Segments) bool) {
		lastOmit := 0

		for i, s := range ss {
			if named, ok := s.(NamedSegment); ok {
				if named.Multiple() {
					if !yield(ss[lastOmit : i+1]) {
						return
					}
					lastOmit = i + 1
				}
			}
		}

		if lastOmit > 0 {
			if lastOmit < len(ss) {
				if !yield(ss[lastOmit:]) {
					return
				}
			}
		} else {
			if !yield(ss[:]) {
				return
			}
		}
	}
}

func (ss Segments) String() string {
	b := &strings.Builder{}
	b.WriteString("/")

	for i, s := range ss {
		if i > 0 {
			b.WriteString("/")
		}
		b.WriteString(s.String())
	}
	return b.String()
}

func (ss Segments) Encode(params map[string]string) string {
	ss1 := make(Segments, len(ss))

	for idx, s := range ss {
		if named, ok := s.(NamedSegment); ok {
			v := params[named.Name()]
			if v == "" {
				v = "-"
			}
			ss1[idx] = segment(v)
			continue
		}
		ss1[idx] = s
	}

	return ss1.String()
}

type Segment interface {
	String() string
}

type NamedSegment interface {
	Segment
	Name() string
	Multiple() bool
}

type segment string

func (s segment) String() string {
	return string(s)
}

var _ NamedSegment = namedSegment{}

type namedSegment struct {
	name     string
	multiple bool
}

func (named namedSegment) Multiple() bool {
	return named.multiple
}

func (named namedSegment) Name() string {
	return named.name
}

func (named namedSegment) String() string {
	if named.multiple {
		return fmt.Sprintf("{%s...}", named.name)
	}
	return fmt.Sprintf("{%s}", named.name)
}

type Matcher interface {
	MatchTo(setter ValueSetter, pathname string) (string, bool)
}

func createMatcher(ss Segments, prefixName string) Matcher {
	n := len(ss)

	m := &matcher{
		prefixName: prefixName,
		Segments:   make(Segments, 0, n),
	}

	for idx, s := range ss {
		m.Segments = append(m.Segments, s)

		// not last
		if idx != n-1 {
			if named, ok := s.(NamedSegment); ok && named.Multiple() {
				return &composedMatcher{
					left:  m,
					right: createMatcher(ss[idx+1:], named.Name()),
				}
			}
		}
	}

	return m
}

type composedMatcher struct {
	left  Matcher
	right Matcher
}

func (c *composedMatcher) String() string {
	return fmt.Sprintf("%s => %s", c.left, c.right)
}

func (c *composedMatcher) MatchTo(setter ValueSetter, pathname string) (string, bool) {
	remain, ok := c.left.MatchTo(setter, pathname)
	if !ok {
		return "", ok
	}
	return c.right.MatchTo(setter, remain)
}

type matcher struct {
	Segments
	prefixName string
}

func (m *matcher) MatchTo(setter ValueSetter, pathname string) (string, bool) {
	parts := splitPath(pathname)

	segN := len(m.Segments)

	if len(parts) < segN {
		return "", false
	}

	strictPrefix := m.prefixName == ""

	offset := 0

	defer func() {
		if m.prefixName != "" {
			setter.Set(m.prefixName, strings.Join(parts[0:offset], "/"))
		}
	}()

	segIdx := 0
	for idx, part := range parts {
		segIdx = idx - offset
		if segIdx >= segN {
			return "", false
		}

		s := m.Segments[segIdx]

		if named, ok := s.(NamedSegment); ok {
			if named.Multiple() {
				remain := strings.Join(parts[idx:], "/")
				setter.Set(named.Name(), remain)
				return remain, true
			}
			setter.Set(named.Name(), part)
			continue
		}

		if s.String() != part {
			if strictPrefix {
				return "", false
			}
			offset++
		}
	}

	// make sure seg all matched
	if segIdx != segN-1 {
		return "", false
	}

	return "", true
}
