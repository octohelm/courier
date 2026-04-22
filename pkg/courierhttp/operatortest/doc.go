// Package operatortest 提供面向单个 courier operator 的 HTTP 测试辅助。
//
// `Serve` 会把一个 operator 包装成临时测试服务器，
// 便于在不组装完整路由树的情况下验证请求解码与响应编码。
package operatortest
