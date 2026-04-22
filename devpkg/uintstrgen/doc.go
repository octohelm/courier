// Package uintstrgen 为无符号整数别名类型生成文本编解码与字符串表示方法。
//
// 生成器名为 `uintstr`，通过类型注释 `+gengo:uintstr` 启用，不额外接收其他参数。
//
// 当前支持的底层类型包括 `uint`、`uint8`、`uint16`、`uint32`、`uint64`。
// 生成结果包含：
//
//   - `UnmarshalText([]byte) error`
//   - `MarshalText() ([]byte, error)`
//   - `String() string`
//
// 其中空文本会被视为零值，零值编码时会返回空文本。
package uintstrgen
