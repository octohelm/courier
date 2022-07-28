package courier

import "context"

type ContextInjector interface {
	InjectContext(ctx context.Context) context.Context
}
