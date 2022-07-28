package org

// +gengo:enum
type Type int

const (
	TYPE_UNKNOWN Type = iota

	TYPE__GOV     // 政府
	TYPE__COMPANY // 企事业单位
)
