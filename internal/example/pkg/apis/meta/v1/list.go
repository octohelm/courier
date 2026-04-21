package v1

// List 表示通用列表返回。
type List[T any] struct {
	// 列表
	Items []*T `json:"items,omitzero"`
	// 总数
	Total int64 `json:"total,omitzero"`
}

func (v *List[T]) Add(item *T) {
	v.Items = append(v.Items, item)
}
