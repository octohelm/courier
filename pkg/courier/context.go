package courier

import (
	"context"
)

// CanInjectContext 表示可注入上下文的接口。
type CanInjectContext interface {
	InjectContext(ctx context.Context) context.Context
}

// ContextInjector 上下文注入器函数类型。
type ContextInjector = func(ctx context.Context) context.Context

// ComposeContextWith 组合多个上下文注入器。
func ComposeContextWith(injectContexts ...ContextInjector) ContextInjector {
	return func(ctx context.Context) context.Context {
		for i := range injectContexts {
			ctx = injectContexts[i](ctx)
		}
		return ctx
	}
}
