package org

import (
	"bytes"
	"fmt"

	"errors"
)

type Type int

const (
	TYPE_UNKNOWN Type = iota

	TYPE__GOV     // 政府
	TYPE__COMPANY // 企事业单位
)

var InvalidType = errors.New("invalid Type")

func (Type) EnumValues() []any {
	return []any{
		TYPE__GOV,
		TYPE__COMPANY,
	}
}
func (v Type) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Type) UnmarshalText(data []byte) error {
	vv, err := ParseTypeFromString(string(bytes.ToUpper(data)))
	if err != nil {
		return err
	}
	*v = vv
	return nil
}

func ParseTypeFromString(s string) (Type, error) {
	switch s {
	case "GOV":
		return TYPE__GOV, nil
	case "COMPANY":
		return TYPE__COMPANY, nil

	default:
		var i Type
		_, err := fmt.Sscanf(s, "UNKNOWN_%d", &i)
		if err == nil {
			return i, nil
		}
		return TYPE_UNKNOWN, InvalidType
	}
}

func (v Type) String() string {
	switch v {
	case TYPE__GOV:
		return "GOV"
	case TYPE__COMPANY:
		return "COMPANY"

	case TYPE_UNKNOWN:
		return "UNKNOWN"
	default:
		return fmt.Sprintf("UNKNOWN_%d", v)
	}
}

func ParseTypeLabelString(label string) (Type, error) {
	switch label {
	case "政府":
		return TYPE__GOV, nil
	case "企事业单位":
		return TYPE__COMPANY, nil

	default:
		return TYPE_UNKNOWN, InvalidType
	}
}

func (v Type) Label() string {
	switch v {
	case TYPE__GOV:
		return "政府"
	case TYPE__COMPANY:
		return "企事业单位"

	default:
		return fmt.Sprint(v)
	}
}
