// Package taggeduniongen 为被 `+gengo:taggedunion` 标记的联合类型生成编解码与类型分发方法。
//
// 生成器名为 `taggedunion`，通过类型注释 `+gengo:taggedunion` 启用。
//
// 目标类型必须包含名为 `Underlying` 的字段，通过其 struct tag 声明判别字段名
// `discriminator:"<字段名>"`。其余字段通过 `mapping:"<判别值>"` 声明变种映射。
// 通过 `+gengo:taggedunion:underlying=<类型名>` 指定底层接口类型。
//
// 生成内容包括：
//
//   - `Discriminator() string` —— 返回判别字段名
//   - `Underlying() any` —— 返回存储的底层值
//   - `Mapping() map[string]any` —— 返回判别值到变体零值的映射
//   - `SetUnderlying(any)` —— 设置底层值
//   - `IsZero() bool` —— 检查底层值为 nil
//   - `MarshalJSON() ([]byte, error)` —— 序列化底层值（值接收器）
//   - `UnmarshalJSON([]byte) error` —— 局部副本模式反序列化
package taggeduniongen
