# 校验与自定义文本类型规范

`pkg/apis/{domain}/{vMajor}` 不只放数据结构，也负责承载公开契约里的校验与文本表示规则。

## validate 放置位置

优先把字段级校验规则定义在 `pkg/apis/{domain}/{vMajor}` 的模型上。

常见写法：

- 字段 tag 使用 `validate:"..."`
- `github.com/octohelm/courier/pkg/validator` 提供通用引擎
- 常用规则由 `github.com/octohelm/courier/pkg/validator/validators` 提供并注册

要使用常见规则，需要先注册 rules provider：

```go
import (
	_ "github.com/octohelm/courier/pkg/validator/validators"
)
```

具体规则不要在 skill 中完整转述，直接查：

```bash
go doc -all github.com/octohelm/courier/pkg/validator/validators
```

例如：

```go
type Pager struct {
	Limit int64 `name:"limit,omitzero" validate:"@int[-1,50] = 10" in:"query"`
}
```

规则建议：

- 让校验跟着公开模型走，而不是散落在 routes 或实现层
- 复用字段或复用类型时，优先复用已有 rule，不重复写多份近似规则
- endpoint 负责暴露 transport 契约，不负责重复定义业务校验

## 类型级校验

当某个命名类型在多个字段上复用时，优先给类型本身定义规则。

可通过实现 `StructTagValidate() string` 绑定默认校验规则：

```go
type OrgName string

func (OrgName) StructTagValidate() string {
	return "@org-name"
}
```

这类能力会被 `github.com/octohelm/courier/pkg/validator` 与 OpenAPI schema 提取流程识别。

接口定义等细节直接查：

```bash
go doc github.com/octohelm/courier/pkg/validator.WithStructTagValidate
```

适用场景：

- 领域命名类型需要在多个模型中复用
- 某个类型天然有稳定格式约束
- 希望校验规则和类型语义绑定，而不是和单个字段绑定

## 自定义 strfmt

当需要定义自定义格式校验时，可注册 format validator provider。

例如：

```go
func init() {
	validator.Register(validator.NewFormatValidatorProvider("org-name", func(format string) validator.Validator {
		return &validators.StringValidator{
			Format:        format,
			MaxLength:     new(uint64(5)),
			Pattern:       regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`),
			PatternErrMsg: "只能包含小写字母，数字和 -，且必须以小写字母或数字开头",
		}
	}))
}
```

建议：

- 自定义 format 名称直接表达领域语义
- format 注册代码应靠近对应 domain type，便于维护
- 对外暴露的是稳定业务语义，不要把内部实现细节放进 format 名称
- 已内置哪些 strfmt、如何扩展，优先直接查 `go doc -all github.com/octohelm/courier/pkg/validator/validators`

## 自定义文本类型

当一个类型需要作为字符串参与请求编解码、query/path/header/cookie 传输或文档抽取时，可实现：

- `encoding.TextMarshaler`
- `encoding.TextUnmarshaler`

例如：

```go
type OrgType int

func (v OrgType) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *OrgType) UnmarshalText(data []byte) error {
	// parse from text
	return nil
}
```

这样做的意义：

- `github.com/octohelm/courier/pkg/content` 可按文本类型处理请求编解码
- transport 中的 query/path/header/cookie 可复用同一套文本表示
- 文档与 schema 提取也能基于命名类型继续补充校验与格式信息

相关能力细节直接查：

```bash
go doc github.com/octohelm/courier/pkg/content
go doc github.com/octohelm/courier/pkg/openapi/jsonschema
```

适用建议：

- 枚举类命名类型优先考虑实现文本编解码
- 需要稳定字符串表示的 ID、类型、状态值也适合使用该方式
- 如果只是内部临时结构，不必为了形式统一强行实现

## 组合建议

一类公开命名类型通常可以同时具备：

- 字符串文本表示：`MarshalText` / `UnmarshalText`
- 类型级校验：`StructTagValidate() string`
- 自定义 format 校验：`validator.NewFormatValidatorProvider(...)`

常见组合方式：

1. 用命名类型表达业务语义
2. 用 `MarshalText` / `UnmarshalText` 定义公开文本表示
3. 用 `StructTagValidate()` 绑定默认规则
4. 需要时注册自定义 format validator

这样能让模型、transport、校验和文档抽取保持一致。
