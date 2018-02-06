# Middleware -- context handling and chaining primitives for Golang

## Description

This is a tiny general-purpose function collection for building complex logic
on contexts. The following primitives are provided:

* `Handler` -- function that receives object of type `context.Context` and performs some actions.

* `Middleware` -- function that processes context and defers execution to its second parameter, handler.

* `Predicate` -- function that simply returns `true` or `false` based on context.

* `Factory` -- function builds middlewares from context.

* Middlewares are chained together using `Compose` function. Its result is a new middleware.

* Middleware  can be applied to the `Handler`, resulting in a new handler.

* `Optional` creates from the predicate and collection of middlewares a new middleware that will be executed only when context yields `true` from predicate.

* `Lazy` transforms `Factory` into `Middleware` on demand.

* Middleware can `Use` other middlewares to build further chain or can be `Branch`-ed using predicate and alternative chain of middlewares.

* Handler can be attached to the middleware with `On` method, it will be called when predicate return `true`.

## License

This code is released under
[MIT](https://github.com/andviro/middleware/blob/master/LICENSE) license.
