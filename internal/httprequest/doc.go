// Package httprequest 提供对 `net/http.Request` 的轻量抽象与辅助适配。
//
// 它把 header、body、path 参数读取等常用能力统一成较小接口，
// 供请求解码和 handler 层复用。
package httprequest
