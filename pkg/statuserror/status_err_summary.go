package statuserror

import (
	"bytes"
	"strconv"
	"text/scanner"

	"github.com/pkg/errors"
)

type ErrorWithSummary interface {
	Summary() string
}

func Summary(err error) (string, bool) {
	if s, ok := err.(ErrorWithSummary); ok {
		return s.Summary(), true
	}
	e := FromErr(err)
	if e == nil {
		return "", false
	}
	return e.Summary(), true
}

func ParseStatusErrSummary(str string) (*StatusErr, error) {
	s := &scanner.Scanner{}
	s.Init(bytes.NewBufferString(str))

	err := &StatusErr{}

	key := ""

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		txt := s.TokenText()

		switch txt {
		case "=":
		case ",", "}":
			key = ""
		default:
			switch key {
			case "key":
				err.Key, _ = strconv.Unquote(txt)
			case "code":
				i, _ := strconv.ParseInt(txt, 10, 64)
				err.Code = int(i)
			case "msg":
				err.Msg, _ = strconv.Unquote(txt)
			case "canBeTalkError":
				err.CanBeTalkError = true
			default:
				key = txt
			}
		}
	}

	if err.Key == "" {
		return nil, errors.Errorf("invalid status err summary: %s", s)
	}

	return err, nil
}
