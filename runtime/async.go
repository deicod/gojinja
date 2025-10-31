package runtime

// Awaitable represents a value that can be awaited inside templates when async
// support is enabled. Implementations should perform any deferred work and
// return the resulting value that should be exposed to the template.
type Awaitable interface {
	Await(ctx *Context) (interface{}, error)
}

// SimpleAwaitable mirrors Awaitable but does not receive rendering context.
// This allows lightweight awaitables that only need to return a value and
// optional error.
type SimpleAwaitable interface {
	Await() (interface{}, error)
}
