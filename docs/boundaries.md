# 手写、约定与生成的边界

`courier` 既不是“全部手写”的库，也不是“全部交给生成器”的库。更实用的做法是明确哪些部分建议手写，哪些部分交给约定，哪些部分适合交给生成器。

## operator

建议手写：

- 业务相关的 `Output` 逻辑
- 真正影响语义的输入输出字段

建议交给约定：

- 方法和路径的 tag
- 参数来自 path、query、body 的声明

建议交给生成：

- 路由注册
- 部分响应描述
- 部分错误响应描述

参考：

- [`internal/example/pkg/endpoints`](https://github.com/octohelm/courier/tree/main/internal/example/pkg/endpoints)
- [`internal/example/cmd/example/routes`](https://github.com/octohelm/courier/tree/main/internal/example/cmd/example/routes)

## response

建议手写：

- 业务响应对象本身
- 必要时显式写出的特殊响应行为

建议交给约定：

- 通过返回值类型表达常规响应形状

建议交给生成：

- 能稳定从 `Output` 推导出的 `ResponseContent`
- 能稳定从返回链路识别的错误描述

## injectable

建议手写：

- service 接口
- route/operator 对 service 的真实调用逻辑

建议交给约定：

- struct 字段上的 `inject` / `provide` tag

建议交给生成：

- `FromContext`
- `InjectContext`
- `Init`

## client

建议手写：

- 业务上层的调用编排
- 特定场景下的 transport 或认证包装

建议交给约定：

- 请求模型和响应模型的结构

建议交给生成：

- 基于 OpenAPI 的 typed client

## OpenAPI

建议手写：

- 必须显式定义的契约信息
- 非常特殊的文档扩展点

建议交给约定：

- 从类型、tag 和 operator 能稳定推导出的参数与响应

建议交给生成或扫描：

- 常规 schema
- 常规 operation 文档

## 一个简单判断

可以用下面这个经验法则：

- 影响业务语义的，优先手写
- 能从稳定约定直接推导的，优先交给约定
- 重复样板已经显现的，优先交给生成

如果团队开始频繁讨论“这段代码是不是又重复了”，通常就到了该引入生成器的时机。
