// Package transformers 注册 `pkg/content` 使用的内置内容转换器。
//
// 当前实现覆盖 JSON、text、octet-stream、multipart/form-data 与
// application/x-www-form-urlencoded 等常见媒体类型。
//
// 导入本包通常不需要显式调用入口；各转换器会在 init 期间向内容层注册。
package transformers
