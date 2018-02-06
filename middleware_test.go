package middleware_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	mw "bitbucket.org/andviro/pif/telegram/middleware"
)

func handlerFactory(s string, dest *bytes.Buffer) mw.Handler {
	return func(ctx context.Context) {
		fmt.Fprintf(dest, "h(%s)\n", s)
	}
}

func mwFactory(s string, dest *bytes.Buffer) mw.Middleware {
	return func(ctx context.Context, next mw.Handler) {
		fmt.Fprintf(dest, "mw(%s)\n", s)
		next(ctx)
	}
}

func Test_Use_Then(t *testing.T) {
	buf := new(bytes.Buffer)
	h1 := handlerFactory("1", buf)
	mw1 := mwFactory("1", buf)
	mw2 := mwFactory("2", buf)
	h2 := h1.Use(mw1, mw2)
	h2(context.TODO())
	if buf.String() != "mw(1)\nmw(2)\nh(1)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
	buf.Reset()
	h3 := mw1.Then(h1)
	h3(context.TODO())
	if buf.String() != "mw(1)\nh(1)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
}

func Test_On(t *testing.T) {
	buf := new(bytes.Buffer)
	ctx1 := context.WithValue(context.TODO(), "key", 1)
	ctx2 := context.WithValue(context.TODO(), "key", 2)
	ctx3 := context.WithValue(context.TODO(), "key", 3)
	h1 := handlerFactory("1", buf)
	h2 := handlerFactory("2", buf)
	p1 := func(ctx context.Context) bool {
		return ctx.Value("key").(int) == 1
	}
	p2 := func(ctx context.Context) bool {
		return ctx.Value("key").(int) == 2
	}
	m := mw.Safe.
		On(p1, h1).
		On(p2, h2)
	m(ctx1, nil)
	if buf.String() != "h(1)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
	buf.Reset()
	m(ctx2, nil)
	if buf.String() != "h(2)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
	buf.Reset()
	m(ctx3, nil)
	if buf.String() != "" {
		t.Errorf("unexpected: %q", buf.String())
	}
}

func Test_Use(t *testing.T) {
	buf := new(bytes.Buffer)
	ctx := context.TODO()
	mw1 := mwFactory("1", buf)
	mw2 := mwFactory("2", buf)
	mw3 := mwFactory("3", buf)
	m := mw.Safe.Use(mw1)
	m(ctx, nil)
	if buf.String() != "mw(1)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
	m = m.Use(mw2, mw3)
	buf.Reset()
	m(ctx, nil)
	if buf.String() != "mw(2)\nmw(3)\nmw(1)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
}

func Test_Compose(t *testing.T) {
	buf := new(bytes.Buffer)
	ctx := context.TODO()
	mw1 := mwFactory("1", buf)
	mw2 := mwFactory("2", buf)
	mw3 := mwFactory("3", buf)
	m := mw.Compose(mw1, mw2, mw3, mw.Safe)
	m(ctx, nil)
	if buf.String() != "mw(1)\nmw(2)\nmw(3)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
}

func Test_Optional(t *testing.T) {
	buf := new(bytes.Buffer)
	ctx1 := context.WithValue(context.TODO(), "key", 1)
	ctx2 := context.WithValue(context.TODO(), "key", 2)
	h1 := handlerFactory("1", buf)
	m1 := mwFactory("1", buf)
	p1 := func(ctx context.Context) bool {
		return ctx.Value("key").(int) == 1
	}
	m := mw.Optional(p1, m1)
	m(ctx2, h1)
	if buf.String() != "h(1)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
	buf.Reset()
	m(ctx1, nil)
	if buf.String() != "mw(1)\n" {
		t.Errorf("unexpected: %q", buf.String())
	}
}
