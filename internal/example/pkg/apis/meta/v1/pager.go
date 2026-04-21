package v1

// Pager 表示通用分页参数。
type Pager struct {
	// 偏移量
	Offset int64 `name:"offset,omitzero" in:"query"`
	// 单次拉取条数，默认 10，最大 50
	Limit int64 `name:"limit,omitzero" validate:"@int[-1,50] = 10" in:"query"`
}
