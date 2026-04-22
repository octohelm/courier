// Package service 定义示例组织域的服务接口，并作为路由层与具体实现之间的注入边界。
//
// 这个包只保留稳定接口，不负责具体存储；内存实现位于同级 `mem` 子包。
//
//go:generate go tool gen .
package service
