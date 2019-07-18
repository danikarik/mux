package mux_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danikarik/mux"
	gorillamux "github.com/gorilla/mux"
)

type contextKey string

const userID = contextKey("user_id")

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

func middlewareContextFunc(id int) mux.MiddlewareFunc {
	return func(w http.ResponseWriter, r *http.Request) (context.Context, error) {
		ctx := context.WithValue(r.Context(), userID, id)
		return ctx, nil
	}
}

func middlewareContextFuncForBypass(id int) gorillamux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), userID, id)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func TestMiddlewareContext(t *testing.T) {
	testCases := []struct {
		Name        string
		Middlewares []mux.MiddlewareFunc
		Code        int
		Expected    string
	}{
		{
			Name: "Single",
			Middlewares: []mux.MiddlewareFunc{
				middlewareContextFunc(1),
			},
			Code:     http.StatusOK,
			Expected: "1",
		},
		{
			Name: "Multiple",
			Middlewares: []mux.MiddlewareFunc{
				middlewareContextFunc(1),
				middlewareContextFunc(2),
				middlewareContextFunc(3),
			},
			Code:     http.StatusOK,
			Expected: "3",
		},
		{
			Name: "Error",
			Middlewares: []mux.MiddlewareFunc{
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

func TestMiddlewareContextBypass(t *testing.T) {
	testCases := []struct {
		Name        string
		Middlewares []gorillamux.MiddlewareFunc
		Code        int
		Expected    string
	}{
		{
			Name: "Single",
			Middlewares: []gorillamux.MiddlewareFunc{
				middlewareContextFuncForBypass(1),
			},
			Code:     http.StatusOK,
			Expected: "1",
		},
		{
			Name: "Multiple",
			Middlewares: []gorillamux.MiddlewareFunc{
				middlewareContextFuncForBypass(1),
				middlewareContextFuncForBypass(2),
				middlewareContextFuncForBypass(3),
			},
			Code:     http.StatusOK,
			Expected: "3",
		},
		{
			Name: "Error",
			Middlewares: []gorillamux.MiddlewareFunc{
				middlewareContextFuncForBypass(1),
				func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						code := http.StatusInternalServerError
						http.Error(w, http.StatusText(code), code)
						return
					})
				},
			},
			Code:     http.StatusInternalServerError,
			Expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := mux.NewRouter()
			mux.UseBypass(tc.Middlewares...)
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
