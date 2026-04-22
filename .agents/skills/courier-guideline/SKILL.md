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

## 使用顺序

1. 先确认这次改动落在哪一层：`apis`、`endpoints`、`routes`、client / test。
2. 只打开当前任务需要的 reference，不在 `SKILL.md` 重复抄规则细节。
3. 落地代码时保持“契约定义”和“实现组装”分离。
4. 收尾时确认生成入口、测试层次和引用路径没有漂移。

## 读取导航

- 需要确认目录职责、命名方式、推荐推进顺序时，读 [references/layout.spec.md](references/layout.spec.md)
- 需要确认 `doc.go` 的包级 `gengo` 标签时，读 [references/docgo.spec.md](references/docgo.spec.md)
- 需要定义 domain errors 的放置位置、命名与暴露方式时，读 [references/errors.spec.md](references/errors.spec.md)
- 需要定义 `pkg/apis/{domain}/{vMajor}` 中的
  validate、类型级校验规则和自定义文本类型时，读 [references/validation.spec.md](references/validation.spec.md)
- 需要定义 endpoint、确认 `in`/`name`/`mime`、`ResponseData()`、`ResponseErrors()`
  时，读 [references/endpoints.spec.md](references/endpoints.spec.md)
- 需要编写 `+gengo:injectable` provider、wrapper 注入点、生成代码边界时，读 [references/injectable.spec.md](references/injectable.spec.md)
- 需要处理 client 调用方式、实现层测试或 API 回路测试时，读 [references/testing.md](references/testing.md)

## 完成标准

- 当前改动的代码层次清楚，没有把契约、组装、实现混写。
- 所引用的规则都能在 `references/` 中直接找到，且相对路径有效。
- 若涉及生成或测试，已回到仓库现有入口验证，而不是发明零散流程。
