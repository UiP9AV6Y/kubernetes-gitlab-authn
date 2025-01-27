package tracing

import (
	"context"
)

type requestContextKey int

const requestKey requestContextKey = 0

// NewContextWithRequestIdentifier returns a new Context
// that carries the given request identifier.
func NewContextWithRequestIdentifier(ctx context.Context, rid string) context.Context {
	return context.WithValue(ctx, requestKey, rid)
}

// RequestIdentifierFromContext returns the request identifier value stored in ctx,
// or an empty string if none exists.
func RequestIdentifierFromContext(ctx context.Context) string {
	rid, ok := ctx.Value(requestKey).(string)
	if ok {
		return rid
	}

	return ""
}
