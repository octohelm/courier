// Package courier 定义 courier 的核心抽象，包括 Operator、Router、Client、
// Transport 与结果类型等最小协作接口。
//
// 这个包只描述与协议无关的路由和调用模型；HTTP 相关适配位于 `pkg/courierhttp`。
// 构建服务端路由时通常先用 `NewRouter` 组织操作符，再交给上层传输层适配；
// 构建客户端时则通过 `Client` 和 `Transport` 完成请求发送与结果解码。
//
// +gengo:runtimedoc=false
package courier
