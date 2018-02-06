package middleware

import "context"

// Predicate performs boolean test on the context
type Predicate func(context.Context) bool

// Factory builds new middleware from the context
type Factory func(context.Context) Middleware

// Handler is applied to context
type Handler func(context.Context)

// Use applies middlewares to the handler
func (h Handler) Use(mws ...Middleware) Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i].Then(h)
	}
	return h
}

// Middleware performs some operation on context and delegates execution to the
// next handler
type Middleware func(context.Context, Handler)

// Safe to apply on nil handler
var Safe Middleware = func(ctx context.Context, h Handler) {
	if h != nil {
		h(ctx)
	}
}

// Compose middlewares into one
func Compose(mws ...Middleware) Middleware {
	if len(mws) == 0 {
		return Safe
	}
	if len(mws) == 1 {
		return mws[0]
	}
	return func(ctx context.Context, next Handler) {
		next.Use(mws...)(ctx)
	}
}

// Optional returns safe middleware chain that's applied only if predicate is
// true
func Optional(p Predicate, mws ...Middleware) Middleware {
	return Safe.Branch(p, Compose(mws...))
}

// Lazy produces middleware on demand using factory function and current
// context
func Lazy(f Factory) Middleware {
	return func(ctx context.Context, next Handler) {
		f(ctx)(ctx, next)
	}
}

// Then applies middleware to the handler
func (mw Middleware) Then(h Handler) Handler {
	return func(ctx context.Context) {
		mw(ctx, h)
	}
}

// Use prepends provided middlewares to the current one
func (m Middleware) Use(mws ...Middleware) Middleware {
	return Compose(append(mws, m)...)
}

// Branch constructs conditional middleware chain using predicate
func (mw Middleware) Branch(p Predicate, next Middleware) Middleware {
	return Lazy(func(ctx context.Context) Middleware {
		if p(ctx) {
			return mw.Use(next)
		}
		return mw
	})
}

// On makes sure that handler will be called if predicate is true
func (mw Middleware) On(p Predicate, h Handler) Middleware {
	return func(ctx context.Context, next Handler) {
		if p(ctx) {
			mw(ctx, h)
			return
		}
		mw(ctx, next)
	}
}
