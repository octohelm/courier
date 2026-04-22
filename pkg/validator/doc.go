/*
Package validator 提供基于规则 DSL 的值校验入口。

这个包对外暴露的核心能力很小：

  - 用 Register 注册新的校验器提供者；
  - 用 New 按规则创建 Validator；
  - 调用 Validator.Validate 执行校验。

内置规则由 `pkg/validator/validators` 与 `pkg/validator/strfmt` 在 init 期间注册，
因此正常使用时只需要导入本包即可。

# 规则 DSL

常见规则形式如下：

	// 简单名称
	@name

	// 带参数
	@name<param1>
	@name<param1,param2>

	// 带区间
	@name[from, to)
	@name[length]

	// 带枚举值
	@name{VALUE1,VALUE2,VALUE3}
	@name{%v}

	// 带正则
	@name/\d+/

	// 可选和默认值
	@name?
	@name = value
	@name = 'some string value'

	// 组合规则
	@map<@string[1,10],@string{A,B,C}>
	@map<@string[1,10],@string/\d+/>[0,10]

规则解析后会被转换成具体校验器实例。
*/
package validator
