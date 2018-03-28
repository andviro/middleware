// Package state holds middleware and primitives for simple state machine.
//
package state

import (
	"context"
	"encoding/json"
	"reflect"
	"regexp"

	"github.com/andviro/middleware"
)

type stateKeyType int

const stateKey stateKeyType = iota

// State is represented by Key and passes optional Value
type State struct {
	Key     string
	Payload json.RawMessage
	Prev    *State
}

func (s *State) Error() string {
	return s.Key
}

// GetValue unmarshals state payload into Go value dest
func (s *State) GetValue(dest interface{}) error {
	return json.Unmarshal(s.Payload, dest)
}

// SetValue sets state payload to Go value
func (s *State) SetValue(src interface{}) (err error) {
	s.Payload, err = json.Marshal(src)
	return
}

// Store saves and loads state
type Store interface {
	Get() (*State, error) // read current state value
	Set(*State) error     // store state value
}

// Current extracts current state from the context
func Current(ctx context.Context) (res *State) {
	res, _ = ctx.Value(stateKey).(*State)
	return
}

func lazy(val interface{}) (f func(context.Context) error) {
	switch t := val.(type) {
	case string:
		f = func(context.Context) error { return &State{Key: t} }
	case func(context.Context) error:
		f = t
	case State:
		f = func(context.Context) error { return &t }
	}
	return
}

// Next returns next state or error from handler
func Next(val interface{}) middleware.Middleware {
	f := lazy(val)
	return func(ctx context.Context, next middleware.Handler) (err error) {
		if err = next.Apply(ctx); err != nil {
			return
		}
		return f(ctx)
	}
}

// Push pushes state down on stack
func Push(val interface{}) middleware.Middleware {
	f := lazy(val)
	return Next(func(ctx context.Context) error {
		prev := Current(ctx)
		next, ok := f(ctx).(*State)
		if ok {
			next.Prev = prev
		}
		return next
	})
}

// Pop pops state from stack
func Pop() middleware.Middleware {
	return Next(func(ctx context.Context) error {
		return Current(ctx).Prev
	})
}

// With injects state into context
func With(state State) middleware.Middleware {
	return func(ctx context.Context, next middleware.Handler) error {
		return next.Apply(context.WithValue(ctx, stateKey, state))
	}
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
		if nextState, ok := err.(*State); ok {
			return store.Set(nextState)
		}
		return
	}
}

// Match constructs predicate that matches specified state, state name or
// regular expression
func Match(target interface{}) middleware.Predicate {
	f := func(*State) bool { return false }
	switch t := target.(type) {
	case string:
		f = func(st *State) bool { return t == st.Key }
	case *State:
		f = func(st *State) bool { return reflect.DeepEqual(t, st) }
	case regexp.Regexp:
		f = func(st *State) bool { return t.MatchString(st.Key) }
	}
	return func(ctx context.Context) bool {
		st, ok := ctx.Value(stateKey).(*State)
		return ok && f(st)
	}
}
