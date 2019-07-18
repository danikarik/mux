package mux

import (
	"context"
	"net/http"

	gorillamux "github.com/gorilla/mux"
)

// MiddlewareFunc wraps standard `http.Handler` middleware style with context and error.
type MiddlewareFunc func(w http.ResponseWriter, r *http.Request) (context.Context, error)

// Use appends a MiddlewareFunc to the chain.
func (r *Router) Use(mwf ...MiddlewareFunc) {
	middlewares := []gorillamux.MiddlewareFunc{}
	for _, fn := range mwf {
		middlewares = append(middlewares, r.Wrapper.MiddlewareFunc(fn))
	}
	r.mux.Use(middlewares...)
}

// UseBypass appends a gorilla's `mux.MiddlewareFunc` to the chain.
func (r *Router) UseBypass(mwf ...gorillamux.MiddlewareFunc) {
	r.mux.Use(mwf...)
}
