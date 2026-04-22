// Package errors 定义 validator 使用的结构化错误类型与包装辅助。
//
// 这些错误值会尽量携带字段路径、规则名和失败原因，
// 方便上层统一格式化或转换成 HTTP 错误响应。
package errors
