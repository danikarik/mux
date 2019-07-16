package mux

import (
	"context"
	"net/http"

	gorillamux "github.com/gorilla/mux"
)

func defaultErrorHandler(err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// NewRouter returns a new router instance.
func NewRouter(opts ...func(*Router)) *Router {
	router := &Router{
		mux:          gorillamux.NewRouter(),
		ErrorHandler: defaultErrorHandler,
	}
	for _, opt := range opts {
		opt(router)
	}
	return router
}

// ErrorHandler handles error returned by `Handler`.
type ErrorHandler func(err error, w http.ResponseWriter, r *http.Request)

// Handler wraps standard `http.HandlerFunc` with `error` return value.
type Handler func(w http.ResponseWriter, r *http.Request) error

// Middleware wraps standard `http.Handler` middleware style with context and error.
type Middleware func(w http.ResponseWriter, r *http.Request) (context.Context, error)

// Router wraps `github.com/gorilla/mux` with custom `mux.Handler`.
type Router struct {
	mux          *gorillamux.Router
	ErrorHandler ErrorHandler
}

func (r *Router) serveHandler(fn Handler) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := fn(w, req); err != nil {
			r.ErrorHandler(err, w, req)
		}
	}
}

func (r *Router) handlerFunc(fn Handler) http.HandlerFunc {
	return r.serveHandler(fn)
}

func (r *Router) serveMiddleware(mwf Middleware) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, err := mwf(w, req)
			if err != nil {
				r.ErrorHandler(err, w, req)
				return
			}
			if ctx != nil {
				req = req.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	}
}

func (r *Router) middlewareFunc(mwf Middleware) func(next http.Handler) http.Handler {
	return r.serveMiddleware(mwf)
}

// ServeHTTP dispatches the handler registered in the matched route.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Handle registers a new route with a matcher for the URL path.
func (r *Router) Handle(path string, h http.Handler) {
	r.mux.Handle(path, h)
}

// HandleFunc registers a new route with a matcher for the URL path.
func (r *Router) HandleFunc(path string, h Handler) {
	r.mux.HandleFunc(path, r.handlerFunc(h))
}

// Use appends a MiddlewareFunc to the chain.
func (r *Router) Use(mwf ...Middleware) {
	middlewares := []gorillamux.MiddlewareFunc{}
	for _, fn := range mwf {
		middlewares = append(middlewares, r.middlewareFunc(fn))
	}
	r.mux.Use(middlewares...)
}
