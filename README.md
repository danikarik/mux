# danikarik/mux

Is a small wrapper around [`gorilla/mux`](https://github.com/gorilla/mux) with few additions:

- `r.HandleFunc` accepts custom [`HandlerFunc`](#handlerfunc) or you can use `http.HandlerFunc` using `r.HandleFuncBypass`
- `r.Use` accepts custom [`MiddlewareFunc`](#middlewarefunc) or you can use `func(http.Handler) http.Handler` using `r.UseBypass`
- `NewRouter()` accepts [`Options`](#options)

## HandlerFunc

```go
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error
```

Example:

```go
func userHandler(w http.ResponseWriter, r *http.Request) error {
    id := mux.Vars(r)["id"]
    user, err := loadUser(id)
    if err != nil {
        return err
    }
    return sendJSON(w, user)
}

r := mux.NewRouter()
r.HandleFunc("/", userHandler)
```

## MiddlewareFunc

```go
type MiddlewareFunc func(w http.ResponseWriter, r *http.Request) (context.Context, error)
```

Example:

```go
func authMiddleware(w http.ResponseWriter, r *http.Request) (context.Context, error) {
    sess, err := store.Session(r)
    if err != nil {
        return nil, httpError(http.StatusUnauthorized, "Unauthorized")
    }
    id, err := sess.Value(userIDKey)
    if err != nil {
        return nil, httpError(http.StatusUnauthorized, "Unauthorized")
    }
    return context.WithValue(r.Context(), userIDContextKey, id)
}

r := mux.NewRouter()
r.Use(authMiddleware)
r.HandleFunc("/me", meHandler)
```

## Options

With custom error handler:

```go
func withCustomErrorHandler(r *mux.Router) {
    r.ErrorHandlerFunc = func(err error, w http.ResponseWriter, r *http.Request) {
        switch err.(type) {
        case *HTTPError:
            sendJSON(w, err)
            break
        case *database.Error:
            http.Error(w, err.Message, http.StatusInternalServerError)
        default:
            http.Error(w, err.Error(), http.StatusInternalServerError)
            break
        }
    }
}

r := mux.NewRouter(withCustomErrorHandler)
```

With custom not found handler:

```go
func withNotFound(r *mux.Router) {
    r.NotFoundHandler = func(w http.ResponseWriter, r *http.Request) error {
        // some stuff...
        http.Error(w, "Not Found", http.StatusNotFound)
        return nil
    }
}

r := mux.NewRouter(withNotFound)
```

With custom method not allowed handler:

```go
func withMethodNotAllowed(r *mux.Router) {
    r.MethodNotAllowedHandler = func(w http.ResponseWriter, r *http.Request) error {
        // some stuff...
        http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
        return nil
    }
}

r := mux.NewRouter(withMethodNotAllowed)
```
