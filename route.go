package mux

import (
	"net/http"
	"net/url"

	gorillamux "github.com/gorilla/mux"
)

// Route stores information to match a request and build URLs.
type Route struct {
	route   *gorillamux.Route
	Wrapper Wrapper
}

func routeWithWrapper(wr Wrapper) func(*Route) {
	return func(r *Route) { r.Wrapper = wr }
}

// NewRoute returns a new route instance.
func NewRoute(r *gorillamux.Route, opts ...func(*Route)) *Route {
	route := &Route{route: r}
	for _, opt := range opts {
		opt(route)
	}
	return route
}

// SkipClean reports whether path cleaning is enabled for this route via
// Router.SkipClean.
func (r *Route) SkipClean() bool {
	return r.route.SkipClean()
}

// Match matches the route against the request.
func (r *Route) Match(req *http.Request, match *gorillamux.RouteMatch) bool {
	return r.route.Match(req, match)
}

// MatcherFunc adds a custom function to be used as request matcher.
func (r *Route) MatcherFunc(f gorillamux.MatcherFunc) *Route {
	return NewRoute(r.route.MatcherFunc(f), routeWithWrapper(r.Wrapper))
}

// GetError returns an error resulted from building the route, if any.
func (r *Route) GetError() error {
	return r.route.GetError()
}

// BuildOnly sets the route to never match: it is only used to build URLs.
func (r *Route) BuildOnly() *Route {
	return NewRoute(r.route.BuildOnly(), routeWithWrapper(r.Wrapper))
}

// Handler sets a handler for the route.
func (r *Route) Handler(h http.Handler) *Route {
	return NewRoute(r.route.Handler(h), routeWithWrapper(r.Wrapper))
}

// HandlerFunc sets a handler function for the route.
func (r *Route) HandlerFunc(fn HandlerFunc) *Route {
	return NewRoute(r.route.HandlerFunc(r.Wrapper.HandlerFunc(fn)), routeWithWrapper(r.Wrapper))
}

// HandlerFuncBypass sets a handler function for the route.
func (r *Route) HandlerFuncBypass(fn func(http.ResponseWriter, *http.Request)) *Route {
	return NewRoute(r.route.HandlerFunc(fn), routeWithWrapper(r.Wrapper))
}

// GetHandler returns the handler for the route, if any.
func (r *Route) GetHandler() http.Handler {
	return r.route.GetHandler()
}

// Name sets the name for the route, used to build URLs.
// It is an error to call Name more than once on a route.
func (r *Route) Name(name string) *Route {
	return NewRoute(r.route.Name(name), routeWithWrapper(r.Wrapper))
}

// GetName returns the name for the route, if any.
func (r *Route) GetName() string {
	return r.route.GetName()
}

// Headers adds a matcher for request header values.
func (r *Route) Headers(pairs ...string) *Route {
	return NewRoute(r.route.Headers(pairs...), routeWithWrapper(r.Wrapper))
}

// HeadersRegexp accepts a sequence of key/value pairs, where the value has regex
// support.
func (r *Route) HeadersRegexp(pairs ...string) *Route {
	return NewRoute(r.route.HeadersRegexp(pairs...), routeWithWrapper(r.Wrapper))
}

// Host adds a matcher for the URL host.
func (r *Route) Host(tpl string) *Route {
	return NewRoute(r.route.Host(tpl), routeWithWrapper(r.Wrapper))
}

// Methods adds a matcher for HTTP methods.
func (r *Route) Methods(methods ...string) *Route {
	return NewRoute(r.route.Methods(methods...), routeWithWrapper(r.Wrapper))
}

// Path adds a matcher for the URL path.
func (r *Route) Path(tpl string) *Route {
	return NewRoute(r.route.Path(tpl), routeWithWrapper(r.Wrapper))
}

// PathPrefix adds a matcher for the URL path prefix. This matches if the given
// template is a prefix of the full URL path. See Route.Path() for details on
// the tpl argument.
func (r *Route) PathPrefix(tpl string) *Route {
	return NewRoute(r.route.PathPrefix(tpl), routeWithWrapper(r.Wrapper))
}

// Queries adds a matcher for URL query values.
func (r *Route) Queries(pairs ...string) *Route {
	return NewRoute(r.route.Queries(pairs...), routeWithWrapper(r.Wrapper))
}

// Schemes adds a matcher for URL schemes.
func (r *Route) Schemes(schemes ...string) *Route {
	return NewRoute(r.route.Schemes(schemes...), routeWithWrapper(r.Wrapper))
}

// BuildVarsFunc adds a custom function to be used to modify build variables
// before a route's URL is built.
func (r *Route) BuildVarsFunc(f gorillamux.BuildVarsFunc) *Route {
	return NewRoute(r.route.BuildVarsFunc(f), routeWithWrapper(r.Wrapper))
}

// Subrouter creates a subrouter for the route.
func (r *Route) Subrouter() *Router {
	return NewRouterWithMux(r.route.Subrouter(), func(rt *Router) {
		rt.Wrapper = r.Wrapper
	})
}

// URL builds a URL for the route.
func (r *Route) URL(pairs ...string) (*url.URL, error) {
	return r.route.URL(pairs...)
}

// URLHost builds the host part of the URL for a route. See Route.URL().
func (r *Route) URLHost(pairs ...string) (*url.URL, error) {
	return r.route.URLHost(pairs...)
}

// URLPath builds the path part of the URL for a route. See Route.URL().
func (r *Route) URLPath(pairs ...string) (*url.URL, error) {
	return r.route.URLPath(pairs...)
}

// GetPathTemplate returns the template used to build the route match.
func (r *Route) GetPathTemplate() (string, error) {
	return r.route.GetPathTemplate()
}

// GetPathRegexp returns the expanded regular expression used to match route path.
func (r *Route) GetPathRegexp() (string, error) {
	return r.route.GetPathRegexp()
}

// GetQueriesRegexp returns the expanded regular expressions used to match the
// route queries.
func (r *Route) GetQueriesRegexp() ([]string, error) {
	return r.route.GetQueriesRegexp()
}

// GetQueriesTemplates returns the templates used to build the
// query matching.
func (r *Route) GetQueriesTemplates() ([]string, error) {
	return r.route.GetQueriesTemplates()
}

// GetMethods returns the methods the route matches against.
func (r *Route) GetMethods() ([]string, error) {
	return r.route.GetMethods()
}

// GetHostTemplate returns the template used to build the route match.
func (r *Route) GetHostTemplate() (string, error) {
	return r.route.GetHostTemplate()
}
