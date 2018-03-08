// Package state holds middleware and primitives for simple state machine.
//
package state

import (
	"context"
	"reflect"
	"regexp"

	"github.com/andviro/middleware"
)

type stateKeyType int

const stateKey stateKeyType = iota

// State is represented by Key and passes optional Value
type State struct {
	Key   string
	Value interface{}
}

func (s State) Error() string {
	return s.Key
}

// Store saves and loads state
type Store interface {
	Get() (State, error) // read current state value
	Set(State) error     // store state value
}

// Current extracts current state from the context
func Current(ctx context.Context) (res State) {
	res, _ = ctx.Value(stateKey).(State)
	return
}

// Next returns next state or provided error
func Next(err error, next State) error {
	if err != nil {
		return err
	}
	return next
}

// Machine bulds middleware using provided store factory. Constructed
// middleware injects state taken from store into context and saves next state
// obtained from handler error value.
func Machine(factory func(context.Context) Store) middleware.Middleware {
	return func(ctx context.Context, next middleware.Handler) (err error) {
		store := factory(ctx)
		st, err := store.Get()
		if err != nil {
			return
		}
		err = next(context.WithValue(ctx, stateKey, st))
		if next, ok := err.(State); ok {
			return store.Set(next)
		}
		return
	}
}

// Match constructs predicate that matches specified state, state name or
// regular expression
func Match(target interface{}) middleware.Predicate {
	f := func(State) bool { return false }
	switch t := target.(type) {
	case string:
		f = func(st State) bool { return t == st.Key }
	case State:
		f = func(st State) bool { return reflect.DeepEqual(t, st) }
	case regexp.Regexp:
		f = func(st State) bool { return t.MatchString(st.Key) }
	}
	return func(ctx context.Context) bool {
		st, ok := ctx.Value(stateKey).(State)
		return ok && f(st)
	}
}
