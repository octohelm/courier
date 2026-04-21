// Package injectablegen 为可注入类型生成上下文注入与初始化辅助代码。
//
// 生成器名为 `injectable`，支持两类注释参数：
//
//   - `+gengo:injectable`
//     为 struct 生成 `Init(context.Context) error`，按字段规则从上下文注入依赖。
//   - `+gengo:injectable:provider[=<ProviderType>]`
//     将类型声明为 provider，并生成对应的 `FromContext` / `InjectContext` 辅助函数。
//     当提供值时，会复用指定 provider 类型的 `InjectContext` 入口注入自身。
//
// 对 struct 字段还支持以下参数：
//
//   - `` `inject:""` ``
//     表示该字段需要从 context 中读取 provider。
//   - `` `inject:",opt"` ``
//     可选注入；context 中缺失时不报错。
//   - `` `inject:"-"` ``
//     禁止作为注入字段处理。
//   - `` `provide:""` ``
//     表示该字段会在当前 provider 的 `InjectContext` 中继续写回 context。
//   - `` `provide:"-"` ``
//     禁止作为 provider 字段处理。
//
// 除字段注入外，生成代码还会识别 `beforeInit(context.Context) error` 与
// `afterInit(context.Context) error` 钩子，并在 `Init` 中自动调用。
package injectablegen
