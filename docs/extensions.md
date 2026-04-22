# 关键扩展点说明

`courier` 的核心接口很小，但围绕 `Operator` 叠加了一组小接口来描述生命周期和扩展行为。理解这些接口的“存在原因”，比背名字更重要。

## `CanInit`

定义位置：

- [`pkg/courier/operator.go`](https://github.com/octohelm/courier/blob/main/pkg/courier/operator.go)

作用：

- 在请求参数完成绑定之后、正式执行 `Output` 之前做初始化。

适合：

- 依赖上下文注入的 operator
- 需要在执行前做轻量准备工作的场景

不建议：

- 把复杂业务逻辑塞进 `Init`
- 在 `Init` 里做和输入校验无关的大量副作用

## `OperatorNewer`

作用：

- 自定义 operator 实例的创建方式。

适合：

- 你不希望默认走 `reflect.New(...)`
- 你需要构造一个带预设状态的 operator 实例

不建议：

- 仅为了绕开普通 struct 构造而滥用

## `OperatorInit`

作用：

- 在新实例创建后，把模板 operator 上的状态复制到新实例。

适合：

- operator 本身携带少量初始化配置
- 希望路由注册时的模板对象能把配置传递到运行时实例

## `DefaultsSetter`

作用：

- 在实例创建后统一补默认值。

适合：

- 一些默认参数值不适合靠 tag 或外部注入表达

不建议：

- 把隐蔽的业务默认规则塞进这里，增加可读性负担

## `ContextProvider`

作用：

- 为中间 operator 产出的上下文值指定稳定的 context key。

适合：

- 多个 operator 串联执行，中间结果需要写入 context

不建议：

- 把它当成全局共享状态通道使用

## `OperatorWithoutOutput`

作用：

- 标记该 operator 不作为请求执行链中的可执行输出节点。

常见用途：

- base path、group 这类元信息 operator

example 可参考：

- [`pkg/courierhttp/route.go`](https://github.com/octohelm/courier/blob/main/pkg/courierhttp/route.go)

## 生命周期视角看这些扩展点

从执行顺序看，这些扩展点大致分布在下面阶段：

1. 创建实例：`OperatorNewer`
2. 从模板拷贝状态：`OperatorInit`
3. 补默认值：`DefaultsSetter`
4. 请求绑定后初始化：`CanInit`
5. 中间结果写入上下文：`ContextProvider`

如果只记一句话：

这些小接口不是随机堆出来的，它们都服务于 operator 的实例化、初始化和链式执行。
