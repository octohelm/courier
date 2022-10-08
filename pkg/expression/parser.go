package expression

import (
	"bytes"
	"container/list"
	"strconv"
	"text/scanner"
)

func ParseString(s string) (Expression, error) {
	return Parse([]byte(s))
}

type exScanner struct {
	ex []any
}

func (s *exScanner) Append(arg any) {
	switch x := arg.(type) {
	case *exScanner:
		s.ex = append(s.ex, x)
	case string:
		n, err := strconv.ParseFloat(x, 10)
		if err == nil {
			s.ex = append(s.ex, n)
		} else {
			switch x[0] {
			case '\'', '"':
				unquoted, _ := strconv.Unquote(x)
				s.ex = append(s.ex, unquoted)
			default:
				s.ex = append(s.ex, x)
			}
		}
	}
}

func Parse(b []byte) (Expression, error) {
	s := scanner.Scanner{}
	s.Init(bytes.NewBuffer(b))

	stack := list.New()
	buf := bytes.NewBuffer(nil)
	e := &exScanner{}

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		switch tok {
		case '(':
			fn := &exScanner{ex: []any{buf.String()}}
			buf.Reset()

			e.Append(fn)
			stack.PushBack(fn)
			e = fn
			break
		case ',':
			if buf.Len() > 0 {
				e.Append(buf.String())
				buf.Reset()
			}
		case ')':
			if buf.Len() > 0 {
				e.Append(buf.String())
				buf.Reset()
			}
			stack.Remove(stack.Back())
			if stack.Len() > 0 {
				e = stack.Back().Value.(*exScanner)
			}
		default:
			buf.WriteString(s.TokenText())
		}
	}

	return e.Expression(), nil
}

func (s *exScanner) Expression() Expression {
	if len(s.ex) == 1 {
		if e, ok := s.ex[0].(*exScanner); ok {
			return e.Expression()
		}
	}

	ee := make(Expression, len(s.ex))

	for i := range ee {
		switch x := s.ex[i].(type) {
		case *exScanner:
			ee[i] = x.Expression()
		default:
			ee[i] = x
		}
	}

	return ee
}
