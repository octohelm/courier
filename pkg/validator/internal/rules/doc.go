// Package rules 实现 validator 规则 DSL 的词法与语法解析。
//
// 它负责把形如 `@string[1,10]` 的规则字符串转换成可执行的规则树，
// 供上层校验器提供者消费。
package rules
