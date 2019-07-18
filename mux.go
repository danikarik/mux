package mux

import (
	"net/http"

	gorillamux "github.com/gorilla/mux"
)

func defaultErrorHandler(err error, w http.ResponseWriter, r *http.Request) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// NewRouter returns a new router instance.
func NewRouter(opts ...func(*Router)) *Router {
	router := &Router{
		mux:              gorillamux.NewRouter(),
		ErrorHandlerFunc: defaultErrorHandler,
	}
	for _, opt := range opts {
		opt(router)
	}
	return router.withCustomHandlers()
}

// ErrorHandlerFunc handles error returned by `Handler`.
type ErrorHandlerFunc func(err error, w http.ResponseWriter, r *http.Request)

// HandlerFunc wraps standard `http.HandlerFunc` with `error` return value.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// Router wraps `github.com/gorilla/mux` with custom `mux.Handler`.
type Router struct {
	mux                     *gorillamux.Router
	ErrorHandlerFunc        ErrorHandlerFunc
	NotFoundHandler         HandlerFunc
	MethodNotAllowedHandler HandlerFunc
}

func (r *Router) withCustomHandlers() *Router {
	if r.NotFoundHandler != nil {
		r.mux.NotFoundHandler = r.handlerFunc(r.NotFoundHandler)
	}
	if r.MethodNotAllowedHandler != nil {
		r.mux.MethodNotAllowedHandler = r.handlerFunc(r.MethodNotAllowedHandler)
	}
	return r
}

func (r *Router) serveHandler(fn HandlerFunc) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := fn(w, req); err != nil {
			r.ErrorHandlerFunc(err, w, req)
		}
	}
}

func (r *Router) handlerFunc(fn HandlerFunc) http.HandlerFunc {
	return r.serveHandler(fn)
}

func (r *Router) serveMiddleware(mwf MiddlewareFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx, err := mwf(w, req)
			if err != nil {
				r.ErrorHandlerFunc(err, w, req)
				return
			}
			if ctx != nil {
				req = req.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	}
}

func (r *Router) middlewareFunc(mwf MiddlewareFunc) func(next http.Handler) http.Handler {
	return r.serveMiddleware(mwf)
}

// ServeHTTP dispatches the handler registered in the matched route.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}

// Get returns a route registered with the given name.
func (r *Router) Get(name string) *gorillamux.Route {
	return r.mux.Get(name)
}

// StrictSlash defines the trailing slash behavior for new routes. The initial
// value is false.
func (r *Router) StrictSlash(value bool) *Router {
	r.mux = r.mux.StrictSlash(value)
	return r
}

// SkipClean defines the path cleaning behaviour for new routes. The initial
// value is false.
func (r *Router) SkipClean(value bool) *Router {
	r.mux = r.mux.SkipClean(value)
	return r
}

// UseEncodedPath tells the router to match the encoded original path
// to the routes.
func (r *Router) UseEncodedPath() *Router {
	r.mux = r.mux.UseEncodedPath()
	return r
}

// NewRoute registers an empty route.
func (r *Router) NewRoute() *gorillamux.Route {
	return r.mux.NewRoute()
}

// Name registers a new route with a name.
// See Route.Name().
func (r *Router) Name(name string) *gorillamux.Route {
	return r.NewRoute().Name(name)
}

// Handle registers a new route with a matcher for the URL path.
// See Route.Path() and Route.Handler().
func (r *Router) Handle(path string, h http.Handler) *gorillamux.Route {
	return r.mux.Handle(path, h)
}

// HandleFunc registers a new route with a matcher for the URL path.
// See Route.Path() and Route.HandlerFunc().
func (r *Router) HandleFunc(path string, h HandlerFunc) *gorillamux.Route {
	return r.mux.HandleFunc(path, r.handlerFunc(h))
}

// Headers registers a new route with a matcher for request header values.
// See Route.Headers().
func (r *Router) Headers(pairs ...string) *gorillamux.Route {
	return r.NewRoute().Headers(pairs...)
}

// Host registers a new route with a matcher for the URL host.
// See Route.Host().
func (r *Router) Host(tpl string) *gorillamux.Route {
	return r.NewRoute().Host(tpl)
}

// MatcherFunc registers a new route with a custom matcher function.
// See Route.MatcherFunc().
func (r *Router) MatcherFunc(f gorillamux.MatcherFunc) *gorillamux.Route {
	return r.NewRoute().MatcherFunc(f)
}

// Methods registers a new route with a matcher for HTTP methods.
// See Route.Methods().
func (r *Router) Methods(methods ...string) *gorillamux.Route {
	return r.NewRoute().Methods(methods...)
}

// Path registers a new route with a matcher for the URL path.
// See Route.Path().
func (r *Router) Path(tpl string) *gorillamux.Route {
	return r.NewRoute().Path(tpl)
}

// PathPrefix registers a new route with a matcher for the URL path prefix.
// See Route.PathPrefix().
func (r *Router) PathPrefix(tpl string) *gorillamux.Route {
	return r.NewRoute().PathPrefix(tpl)
}

// Queries registers a new route with a matcher for URL query values.
// See Route.Queries().
func (r *Router) Queries(pairs ...string) *gorillamux.Route {
	return r.NewRoute().Queries(pairs...)
}

// Schemes registers a new route with a matcher for URL schemes.
// See Route.Schemes().
func (r *Router) Schemes(schemes ...string) *gorillamux.Route {
	return r.NewRoute().Schemes(schemes...)
}

// BuildVarsFunc registers a new route with a custom function for modifying
// route variables before building a URL.
func (r *Router) BuildVarsFunc(f gorillamux.BuildVarsFunc) *gorillamux.Route {
	return r.NewRoute().BuildVarsFunc(f)
}

// Walk walks the router and all its sub-routers, calling walkFn for each route
// in the tree. The routes are walked in the order they were added. Sub-routers
// are explored depth-first.
func (r *Router) Walk(walkFn gorillamux.WalkFunc) error {
	return r.mux.Walk(walkFn)
}

// Vars returns the route variables for the current request, if any.
func Vars(r *http.Request) map[string]string {
	return gorillamux.Vars(r)
}

// CurrentRoute returns the matched route for the current request, if any.
func CurrentRoute(r *http.Request) *gorillamux.Route {
	return gorillamux.CurrentRoute(r)
}
