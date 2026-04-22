// Package service 定义示例制品仓库域的服务接口，并作为路由层与具体实现之间的注入边界。
//
// 这个包只描述 blob 与 manifest 的领域操作；内存实现位于同级 `mem` 子包。
//
//go:generate go tool gen .
package service
