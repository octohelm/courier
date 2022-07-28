package courier

import "context"

type ContextInjector = func(ctx context.Context) context.Context

func ComposeContextWith(injectContexts ...ContextInjector) ContextInjector {
	return func(ctx context.Context) context.Context {
		for i := range injectContexts {
			ctx = injectContexts[i](ctx)
		}
		return ctx
	}
}
