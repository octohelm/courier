package statuserror

type DeprecatedErrorField struct {
	In    string `json:"in"`
	Field string `json:"field"`
	Msg   string `json:"msg"`
}
