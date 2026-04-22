// Package httprouter 将 courier 路由树适配为 `net/http` Handler。
//
// 它负责把 `courier.Router` 展开成可执行的 HTTP 路由，串接中间件，
// 并按需暴露 OpenAPI 文档与文档查看页。
package httprouter
