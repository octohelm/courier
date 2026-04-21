// Package operatorgen 为 courier operator 生成路由注册与响应描述辅助代码。
//
// 生成器名为 `operator`，面向导出的 courier operator 类型工作，并支持以下参数：
//
//   - `+gengo:operator:register=<RouterVar>`
//     为当前包生成 `init()` 注册代码，效果等价于
//     `RouterVar.Register(courier.NewRouter(&Operator{}))`。
//     该参数可重复声明，用于注册到多个 router 变量。
//
// 生成内容主要包括：
//
//   - 根据 `Output` 方法推断 `ResponseData`、`ResponseContent`、
//     `ResponseStatusCode`、`ResponseContentType`；
//   - 扫描 `Output` 返回链路中的 `statuserror`，补充 `ResponseErrors`；
//   - 在声明了 `register` 参数时自动输出路由注册入口。
package operatorgen
