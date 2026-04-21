package courier

import (
	"net/url"
)

// FromMetas 从多个元数据创建合并的元数据。
func FromMetas(metas ...Metadata) Metadata {
	m := Metadata{}
	for _, meta := range metas {
		m.Merge(meta)
	}
	return m
}

// Metadata 表示键值对元数据类型。
type Metadata map[string][]string

// String 将元数据转换为查询字符串格式。
func (m Metadata) String() string {
	return url.Values(m).Encode()
}

// Del 删除指定键的元数据。
func (m Metadata) Del(key string) {
	delete(m, key)
}

// Merge 合并其他元数据到当前元数据。
func (m Metadata) Merge(metadata Metadata) {
	for key, values := range metadata {
		m.Set(key, values...)
	}
}

// Add 添加键值对到元数据。
func (m Metadata) Add(key, value string) {
	if values, ok := m[key]; ok {
		m[key] = append(values, value)
	} else {
		m.Set(key, value)
	}
}

// Set 设置指定键的元数据值。
func (m Metadata) Set(key string, values ...string) {
	m[key] = values
}

// Has 检查是否包含指定键。
func (m Metadata) Has(key string) bool {
	_, ok := m[key]
	return ok
}

// Get 获取指定键的第一个值。
func (m Metadata) Get(key string) string {
	if v := m[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}
