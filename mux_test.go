package mux_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danikarik/mux"
)

type contextKey string

const userID = contextKey("user_id")

func errorHandler(code int) func(err error, w http.ResponseWriter, r *http.Request) {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		http.Error(w, err.Error(), code)
	}
}

func newStatusError(got, exp int) *statusError {
	return &statusError{
		Got:      got,
		Expected: exp,
	}
}

type statusError struct {
	Got      int
	Expected int
}

func (e *statusError) Error() string {
	return fmt.Sprintf("got wrong status code %v, expected %v", e.Got, e.Expected)
}

func okHandler(w http.ResponseWriter, r *http.Request) error {
	code := http.StatusOK
	w.WriteHeader(code)
	w.Write([]byte(http.StatusText(code)))
	return nil
}

func contextHandler(w http.ResponseWriter, r *http.Request) error {
	id, ok := r.Context().Value(userID).(int)
	if !ok {
		return errors.New("wrong context key")
	}
	if id == 0 {
		return errors.New("empty context key")
	}
	message := fmt.Sprintf("%d", id)
	w.Write([]byte(message))
	return nil
}

func failedHandler(w http.ResponseWriter, r *http.Request) error {
	return errors.New("internal error occured")
}

func custom404(w http.ResponseWriter, r *http.Request) error {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusBadRequest)
	return nil
}

func custom405(w http.ResponseWriter, r *http.Request) error {
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusBadRequest)
	return nil
}

func middlewareContextFunc(id int) mux.Middleware {
	return func(w http.ResponseWriter, r *http.Request) (context.Context, error) {
		ctx := context.WithValue(r.Context(), userID, id)
		return ctx, nil
	}
}

func TestHandlerFuncStatusCode(t *testing.T) {
	testCases := []struct {
		Name     string
		Handler  mux.Handler
		Option   func(*mux.Router)
		Expected int
	}{
		{
			Name:     "OK",
			Handler:  okHandler,
			Option:   func(r *mux.Router) { r.ErrorHandler = errorHandler(500) },
			Expected: http.StatusOK,
		},
		{
			Name:     "ServerError",
			Handler:  failedHandler,
			Option:   func(r *mux.Router) { r.ErrorHandler = errorHandler(500) },
			Expected: http.StatusInternalServerError,
		},
		{
			Name:     "CustomError",
			Handler:  failedHandler,
			Option:   func(r *mux.Router) { r.ErrorHandler = errorHandler(400) },
			Expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := mux.NewRouter(tc.Option)
			mux.HandleFunc("/", tc.Handler)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)
			resp := w.Result()

			if resp.StatusCode != tc.Expected {
				err := newStatusError(resp.StatusCode, tc.Expected)
				t.Fatal(err)
			}
		})
	}
}

func TestHandlerStatusCode(t *testing.T) {
	testCases := []struct {
		Name     string
		Handler  http.HandlerFunc
		Expected int
	}{
		{
			Name:     "OK",
			Handler:  func(w http.ResponseWriter, r *http.Request) {},
			Expected: http.StatusOK,
		},
		{
			Name: "ServerError",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			Expected: http.StatusInternalServerError,
		},
		{
			Name: "BadRequest",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			Expected: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := mux.NewRouter()
			mux.Handle("/", http.HandlerFunc(tc.Handler))

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)
			resp := w.Result()

			if resp.StatusCode != tc.Expected {
				err := newStatusError(resp.StatusCode, tc.Expected)
				t.Fatal(err)
			}
		})
	}
}

func TestMiddlewareContext(t *testing.T) {
	testCases := []struct {
		Name        string
		Middlewares []mux.Middleware
		Code        int
		Expected    string
	}{
		{
			Name: "Single",
			Middlewares: []mux.Middleware{
				middlewareContextFunc(1),
			},
			Code:     http.StatusOK,
			Expected: "1",
		},
		{
			Name: "Multiple",
			Middlewares: []mux.Middleware{
				middlewareContextFunc(1),
				middlewareContextFunc(2),
				middlewareContextFunc(3),
			},
			Code:     http.StatusOK,
			Expected: "3",
		},
		{
			Name: "Error",
			Middlewares: []mux.Middleware{
				middlewareContextFunc(1),
				func(w http.ResponseWriter, r *http.Request) (context.Context, error) {
					return nil, errors.New("unauthorized")
				},
			},
			Code:     http.StatusInternalServerError,
			Expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := mux.NewRouter()
			mux.Use(tc.Middlewares...)
			mux.HandleFunc("/", contextHandler)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)
			resp := w.Result()

			if resp.StatusCode != tc.Code {
				err := newStatusError(resp.StatusCode, tc.Code)
				t.Fatal(err)
			}

			if resp.StatusCode == http.StatusOK {
				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()

				if string(data) != tc.Expected {
					err := fmt.Errorf("failed: got %s, expected %s", string(data), tc.Expected)
					t.Fatal(err)
				}
			}
		})
	}

}

func TestCustomHandlers(t *testing.T) {
	testCases := []struct {
		Name     string
		Path     string
		Method   string
		Option   func(*mux.Router)
		Code     int
		Expected string
	}{
		{
			Name:     "200",
			Path:     "/",
			Method:   "GET",
			Option:   func(r *mux.Router) {},
			Code:     http.StatusOK,
			Expected: http.StatusText(http.StatusOK),
		},
		{
			Name:     "404",
			Path:     "/someroute",
			Method:   "GET",
			Option:   func(r *mux.Router) { r.NotFoundHandler = custom404 },
			Code:     http.StatusBadRequest,
			Expected: http.StatusText(http.StatusNotFound),
		},
		{
			Name:     "405",
			Path:     "/",
			Method:   "POST",
			Option:   func(r *mux.Router) { r.MethodNotAllowedHandler = custom405 },
			Code:     http.StatusBadRequest,
			Expected: http.StatusText(http.StatusMethodNotAllowed),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := mux.NewRouter(tc.Option)
			mux.HandleFunc("/", okHandler).Methods("GET")

			r := httptest.NewRequest(tc.Method, tc.Path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, r)
			resp := w.Result()

			if resp.StatusCode != tc.Code {
				err := newStatusError(resp.StatusCode, tc.Code)
				t.Fatal(err)
			}

			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if strings.Trim(string(data), "\n") != tc.Expected {
				err := fmt.Errorf("failed: got %s, expected %s", string(data), tc.Expected)
				t.Fatal(err)
			}
		})
	}
}
