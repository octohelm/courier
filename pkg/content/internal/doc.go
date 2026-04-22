// Package internal 为 `pkg/content` 提供请求编解码、参数映射与内容转换器注册等内部实现。
//
// 这个包不直接面向业务使用；上层通常通过 `pkg/content` 或
// `pkg/courierhttp/transport` 间接使用这里的能力。
package internal
