---
name: courier-guideline
description: 当任务涉及基于 github.com/octohelm/courier 定义 API 契约、组装 routes 或编写 client 调用时使用。
---

# Courier Guideline

用于按 `github.com/octohelm/courier` 的约定组织 API 契约、routes 组装与 client 调用边界。

## 适用范围

- 新增或修改 `pkg/apis/{domain}/{vMajor}`、`pkg/endpoints/{domain}/{vMajor}`、`cmd/*/routes/**`
- 基于 courier 定义请求/响应契约、错误集合、transport 参数位置
- 为 routes 接入可注入依赖，或补基于 endpoints 的 client / 回路测试

不适用：

- 只改纯业务实现且不触及 courier 契约或 routes
- 重写项目自己的实现层目录规范
- 手写或手改生成产物

## 输入

至少需要：

- 目标 Go 包路径或目标目录
- 要新增或修改的 API 域对象
- 现有实现层入口，或可新增的可注入依赖

如果实现层目录尚未固定，保持对实现布局中立，只要求 routes 依赖稳定的可注入 provider。

## 关键约定

- `pkg/apis` 与 `pkg/endpoints` 是契约层；`cmd/*/routes` 是组装层。
- 契约层按 `pkg/apis/{domain}/{vMajor}` 与 `pkg/endpoints/{domain}/{vMajor}` 组织，不把实现逻辑写进 endpoint。
- routes 通过可注入依赖连接实现，不强绑某个固定实现目录。
- 公开校验、错误类型、文本编解码规则优先放在 `pkg/apis/{domain}/{vMajor}`。
- 生成文件例如 `zz_generated.injectable.go` 视为产物，不手写、不手改。

## 三层架构

定义 API 时按以下顺序推进，每层只依赖上一层：

```

apis/{domain}/{v1}   ← 模型、错误、校验
        │
        ▼
endpoints/{domain}/{v1} ← HTTP 契约（路径、参数、返回）
        │
        ▼
routes/{domain}/{v1}    ← 组装 + 注入 + 实现连接
```

1. **apis** — 定义请求/响应模型、错误类型、字段校验规则。不涉及 HTTP。
2. **endpoints** — 嵌入 HTTP method 类型、声明路由 `path` 和参数位置 `in`、声明 `ResponseData()` 和 `ResponseErrors()`。不写业务逻辑。
3. **routes** — 嵌入 endpoint + 注入 service provider、实现 `Output()`。通过 `+gengo:injectable` 让生成器补齐注入代码。

## 使用顺序

1. 先确认改动落在哪一层。
2. 按对应 reference 落代码，保持契约定义和实现组装分离。
3. 需要了解具体 API 签名时，优先 `go doc`，不在 skill 中复制手册。
4. 收尾时确认生成入口和测试没有漂移。

## 读取导航

按层选择：

- 整体目录约定 → [references/layout.spec.md](references/layout.spec.md)
- `doc.go` 的 gengo 标签 → [references/docgo.spec.md](references/docgo.spec.md)
- **apis 层** → [references/validation.spec.md](references/validation.spec.md)（校验与自定义类型）、[references/errors.spec.md](references/errors.spec.md)（错误定义）
- **endpoints 层** → [references/endpoints.spec.md](references/endpoints.spec.md)（HTTP 契约写法）
- **routes 层** → [references/injectable.spec.md](references/injectable.spec.md)（注入依赖与组装）
- client / 测试 → [references/testing.md](references/testing.md)

## 完成标准

- 当前改动的代码层次清楚，没有把契约、组装、实现混写。
- 所引用的规则都能在 `references/` 中直接找到，且相对路径有效。
- 若涉及生成或测试，已回到仓库现有入口验证，而不是发明零散流程。
