// Package handler 提供 HTTP handler 级别的基础辅助能力。
//
// 目前主要包含 middleware 组合与路径参数读取上下文封装，
// 供 `pkg/courierhttp/handler/httprouter` 和测试代码复用。
package handler
