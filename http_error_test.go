package mux_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/danikarik/mux"
)

func ExampleHTTPError() {
	err := mux.NewHTTPError(http.StatusBadRequest, "Bad Request")
	data, _ := json.Marshal(err)
	fmt.Println(string(data))

	opts := []func(e *mux.HTTPError){
		func(e *mux.HTTPError) { e.ShowError = true },
		func(e *mux.HTTPError) { e.InternalError = errors.New("Unexpected Error") },
		func(e *mux.HTTPError) { e.InternalMessage = "Failed" },
		func(e *mux.HTTPError) { e.ErrorID = "123" },
	}
	err = mux.NewHTTPError(http.StatusInternalServerError, "Server Error", opts...)
	data, _ = json.Marshal(err)
	fmt.Println(string(data))

	err = mux.NewHTTPError(http.StatusUnauthorized, "Unauthorized").
		WithInternalMessage("Failed").
		WithErrorID("456").
		WithShowError(true)
	data, _ = json.Marshal(err)
	fmt.Println(string(data))

	// Output:
	// {"code":400,"message":"Bad Request"}
	// {"code":500,"message":"Server Error","internalError":"Unexpected Error","internalMessage":"Failed","id":"123"}
	// {"code":401,"message":"Unauthorized","internalMessage":"Failed","id":"456"}
}
