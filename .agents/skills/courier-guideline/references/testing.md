# Client 与测试建议

`github.com/octohelm/courier/pkg/courier.DoWith` 不只用于测试，也可作为常规 client 调用入口。

详情以包文档为准，先查：

```bash
go doc github.com/octohelm/courier/pkg/courier.DoWith
```

优先分两层测试：

1. 实现层测试
2. 请求回路测试

实现层测试：

- 直接覆盖项目自己的实现行为
- 重点覆盖主要业务分支、错误语义、边界条件

请求回路测试：

- 启动本地 `httptest.Server`
- 注入测试实现
- 使用 endpoints + client 走完整请求回路

client 与测试建议：

- 统一优先使用 `github.com/octohelm/courier/pkg/courier.DoWith(ctx, client, op)`
- 当 `ResponseData()` 为 `io.ReadCloser` 时，按响应体直传语义使用返回值
- 对 `io.ReadCloser` 返回值，调用方负责在读取完成后关闭
- 测试中要显式覆盖读取结果与关闭行为
- 局部单测如果只验证某个 handler/operator，可自定义最小 operator，不必引入整套应用入口
