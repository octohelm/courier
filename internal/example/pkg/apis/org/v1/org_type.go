package v1

import (
	"bytes"
	"errors"
	"fmt"
)

type OrgType int

const (
	ORG_TYPE_UNKNOWN OrgType = iota

	ORG_TYPE__GOV     // 政府
	ORG_TYPE__COMPANY // 企事业单位
)

var InvalidOrgType = errors.New("无效的组织类型")

func (OrgType) EnumValues() []any {
	return []any{
		ORG_TYPE__GOV,
		ORG_TYPE__COMPANY,
	}
}

func (v OrgType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OrgType) UnmarshalText(data []byte) error {
	vv, err := parseTypeFromString(string(bytes.ToUpper(data)))
	if err != nil {
		return err
	}
	*v = vv
	return nil
}

func (v OrgType) String() string {
	switch v {
	case ORG_TYPE__GOV:
		return "GOV"
	case ORG_TYPE__COMPANY:
		return "COMPANY"
	case ORG_TYPE_UNKNOWN:
		return "UNKNOWN"
	default:
		return fmt.Sprintf("UNKNOWN_%d", v)
	}
}

func parseTypeFromString(s string) (OrgType, error) {
	switch s {
	case "GOV":
		return ORG_TYPE__GOV, nil
	case "COMPANY":
		return ORG_TYPE__COMPANY, nil
	default:
		var i OrgType
		_, err := fmt.Sscanf(s, "UNKNOWN_%d", &i)
		if err == nil {
			return i, nil
		}
		return ORG_TYPE_UNKNOWN, InvalidOrgType
	}
}
