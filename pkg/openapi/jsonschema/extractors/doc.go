// Package extractors 从 Go 类型与校验规则中提取 JSON Schema 描述。
//
// 这里的提取器会结合反射、validator 规则与运行时文档信息，
// 生成 `pkg/openapi/jsonschema` 中定义的 schema 对象。
package extractors
