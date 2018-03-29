package middleware

import "context"

// Predicate performs boolean test on the context
type Predicate func(context.Context) bool

// Factory builds new middleware from the context
type Factory func(context.Context) Middleware

// Handler is applied to context
type Handler func(context.Context) error

// Default is just an empty pass-through middleware
var Default Middleware

// Apply calls handler on the context. It's safe to use with nil handler.
func (h Handler) Apply(ctx context.Context) error {
	if h == nil {
		return nil
	}
	return h(ctx)
}

// Use applies middlewares to the handler
func (h Handler) Use(mws ...Middleware) Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i].Then(h)
	}
	return h
}

// Middleware performs some operation on context and delegates execution to the
// next handler
type Middleware func(context.Context, Handler) error

// Compose middlewares into one
func Compose(mws ...Middleware) Middleware {
	return func(ctx context.Context, next Handler) error {
		return next.Use(mws...).Apply(ctx)
	}
}

// Optional returns new middleware chain that's applied only if predicate is true
func Optional(p Predicate, mws ...Middleware) Middleware {
	return Middleware(nil).Branch(p, mws...)
}

// Lazy produces middleware on demand using factory function and current
// context
func Lazy(f Factory) Middleware {
	return func(ctx context.Context, next Handler) error {
		return f(ctx).Then(next).Apply(ctx)
	}
}

// Then applies middleware to the handler. If called on nil middleware, returns
// the handler itself.
func (mw Middleware) Then(h Handler) Handler {
	if mw == nil {
		return h
	}
	return func(ctx context.Context) error {
		return mw(ctx, h)
	}
}

// Use prepends provided middlewares to the current one
func (mw Middleware) Use(mws ...Middleware) Middleware {
	return func(ctx context.Context, next Handler) error {
		return mw.Then(next.Use(mws...)).Apply(ctx)
	}
}

// Branch constructs conditional middleware chain using predicate
func (mw Middleware) Branch(p Predicate, mws ...Middleware) Middleware {
	trueMw := mw.Use(mws...)
	return Lazy(func(ctx context.Context) Middleware {
		if p(ctx) {
			return trueMw
		}
		return mw
	})
}

// On makes sure that handler will be called if predicate is true
func (mw Middleware) On(p Predicate, h Handler) Middleware {
	return func(ctx context.Context, next Handler) error {
		if p(ctx) {
			return mw.Then(h).Apply(ctx)
		}
		return mw.Then(next).Apply(ctx)
	}
}

// And constructs predicate as logical AND of its arguments
func And(ps ...Predicate) Predicate {
	return func(ctx context.Context) bool {
		for _, p := range ps {
			if !p(ctx) {
				return false
			}
		}
		return true
	}
}

// Or constructs predicate as logical OR of its arguments
func Or(ps ...Predicate) Predicate {
	return func(ctx context.Context) bool {
		for _, p := range ps {
			if p(ctx) {
				return true
			}
		}
		return false
	}
}

// Not negates predicate
func Not(p Predicate) Predicate {
	return func(ctx context.Context) bool {
		return !p(ctx)
	}
}
