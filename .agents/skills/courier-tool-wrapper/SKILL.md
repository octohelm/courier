---
name: courier-tool-wrapper
description: 当任务涉及基于 github.com/octohelm/courier 定义 API 契约、组装 routes 或编写 client 调用时使用。
---

# Courier Tool Wrapper

按 `github.com/octohelm/courier` 的推荐方式组织 API 契约与 routes 组装，避免把契约定义、服务组装和 client 调用混写在一起。

## 目标

- 用 `pkg/{apis,endpoints}/{domain}/{vMajor}` 组织契约层
- 用 `cmd/*/routes` 负责服务组装和 client 调用对接
- 用可注入依赖连接 routes 与具体实现
- 用 endpoints + client 做请求回路验证

## 输入

至少需要：

- 目标 Go 包路径或目标目录
- 要新增或修改的 API 域对象
- 现有实现层入口，或可新增的可注入依赖

如果实现层目录尚未固定，不要自行发明目录规范；保持对实现布局中立，只要求 routes 依赖可注入依赖。

## 执行流程

1. 先确定契约边界。
   先看 `pkg/{apis,endpoints}` 应承载什么。
2. 再定义 endpoint。
   明确方法、路径、参数位置、成功返回和错误集合。
3. 再组装 wrapper。
   在 `cmd/*/routes` 中嵌入 endpoint，完成服务组装和 client 调用对接。
4. 最后补测试。
   普通接口优先补请求回路测试；实现层若有稳定行为，再补对应实现测试。

## 读取导航

- 需要确认目录职责、命名方式、推荐推进顺序时，读 [references/layout.spec.md](references/layout.spec.md)
- 需要确认 `doc.go` 的包级 `gengo` 标签时，读 [references/docgo.spec.md](references/docgo.spec.md)
- 需要定义 domain errors 的放置位置、命名与暴露方式时，读 [references/errors.spec.md](references/errors.spec.md)
- 需要定义 `pkg/apis/{domain}/{vMajor}` 中的 validate、类型级校验规则和自定义文本类型时，读 [references/validation.spec.md](references/validation.spec.md)
- 需要定义 endpoint、确认 `in`/`name`/`mime`、`ResponseData()`、`ResponseErrors()` 时，读 [references/endpoints.spec.md](references/endpoints.spec.md)
- 需要编写 `+gengo:injectable` provider、wrapper 注入点、生成代码边界时，读 [references/injectable.spec.md](references/injectable.spec.md)
- 需要处理 client 调用方式、实现层测试或 API 回路测试时，读 [references/testing.md](references/testing.md)

## 完成标准

- `pkg/{apis,endpoints}` 保持为契约层
- 契约层按 `pkg/apis/{domain}/{vMajor}` 与 `pkg/endpoints/{domain}/{vMajor}` 组织
- `pkg/endpoints` 不实现 `Output`
- `cmd/*/routes` 负责服务组装和 client 调用对接
- routes 只依赖可注入依赖，不强绑具体实现目录
- 生成文件如 `zz_generated.injectable.go` 视为产物，不手写、不手改
- 文档引用使用真实相对路径
- 仅把需要立即执行的信息保留在 `SKILL.md`
- 不把某个仓库里的示例目录写成共享 skill 的默认项目结构
