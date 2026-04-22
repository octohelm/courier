// Package client 提供基于 courier 协议模型的 HTTP 客户端实现。
//
// `Client` 负责把 operator 请求编码成 HTTP 请求，发送后再按 courier 约定解码
// 成成功结果或 `statuserror`。包内还提供 HTTP transport 链、默认连接策略、
// host alias 与 context 注入等辅助能力。
package client
