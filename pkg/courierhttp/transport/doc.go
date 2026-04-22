// Package transport 提供 courier 与 HTTP 之间的传输层适配。
//
// `NewOutgoingTransport` 用于把请求结构编码成 HTTP 请求，
// `NewIncomingTransport` 用于把 HTTP 请求解码到输入结构并把结果写回响应。
// `Upgrader` 则为需要接管底层连接的场景保留扩展点。
package transport
