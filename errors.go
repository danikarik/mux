package mux

import "net/http"

// ErrorHandlerFunc handles error returned by `Handler`.
type ErrorHandlerFunc func(err error, w http.ResponseWriter, r *http.Request)

func basicErrorFunc(err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// Wrapper defines route wrapping methods and error handling
// for `Router` and `Route`.
type Wrapper interface {
	ServeHandler(HandlerFunc) func(http.ResponseWriter, *http.Request)
	HandlerFunc(HandlerFunc) http.HandlerFunc
	ServeMiddleware(MiddlewareFunc) func(http.Handler) http.Handler
	MiddlewareFunc(MiddlewareFunc) func(http.Handler) http.Handler
	HandleError(error, http.ResponseWriter, *http.Request)
}

// NewDefaultWrapper returns a new default wrapper.
func NewDefaultWrapper(fn ErrorHandlerFunc) Wrapper { return &defaultWrapper{fn} }

type defaultWrapper struct{ ErrorHandler ErrorHandlerFunc }

func (wr *defaultWrapper) ServeHandler(fn HandlerFunc) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			wr.HandleError(err, w, r)
		}
	}
}

func (wr *defaultWrapper) HandlerFunc(fn HandlerFunc) http.HandlerFunc {
	return wr.ServeHandler(fn)
}

func (wr *defaultWrapper) ServeMiddleware(mwf MiddlewareFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := mwf(w, r)
			if err != nil {
				wr.HandleError(err, w, r)
				return
			}
			if ctx != nil {
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (wr *defaultWrapper) MiddlewareFunc(mwf MiddlewareFunc) func(next http.Handler) http.Handler {
	return wr.ServeMiddleware(mwf)
}

func (wr *defaultWrapper) HandleError(err error, w http.ResponseWriter, r *http.Request) {
	wr.ErrorHandler(err, w, r)
}
