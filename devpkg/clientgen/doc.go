// Package clientgen 根据 OpenAPI 文档生成 courier 客户端操作类型与相关数据类型。
//
// 生成器名为 `client`，通常通过类型注释上的 `+gengo:client:*` 参数驱动：
//
//   - `+gengo:client:openapi=<uri>`
//     指定 OpenAPI 文档地址。当前实现要求提供 HTTP 或 HTTPS 地址。
//   - `+gengo:client:typegen-policy=<policy>`
//     控制 schema 到 Go 类型的生成策略，可选值为 `All`、`GoVendorAll`、
//     `GoVendorImported`。默认值为 `GoVendorImported`。
//   - `+gengo:client:openapi:trim-base-path=<prefix>`
//     生成请求路径前，先从 OpenAPI path 中裁掉指定前缀。
//   - `+gengo:client:openapi:include=<operationId>`
//     按 operationId 白名单生成。可重复声明；未配置时默认生成全部 operation。
//
// 生成结果会基于 OpenAPI 中的参数、请求体与成功响应自动展开：
//
//   - operation 参数会转成 `Parameters` 结构体字段；
//   - request body 会转成 `in:"body"` 字段；
//   - 2xx 响应会生成 `ResponseData` 返回类型及相关 schema 定义。
package clientgen
