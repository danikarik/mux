package mux

import (
	"encoding/json"
	"errors"
)

// HTTPError holds http error info.
type HTTPError struct {
	Code            int
	Message         string
	InternalError   error
	InternalMessage string
	ErrorID         string
	ShowError       bool
}

type httpError struct {
	Code            int    `json:"code"`
	Message         string `json:"message"`
	InternalError   string `json:"internal_error,omitempty"`
	InternalMessage string `json:"internal_message,omitempty"`
	ErrorID         string `json:"id,omitempty"`
}

// NewHTTPError returns a new instance of `HTTPError`.
func NewHTTPError(code int, message string, opts ...func(e *HTTPError)) *HTTPError {
	err := &HTTPError{
		Code:    code,
		Message: message,
	}
	for _, opt := range opts {
		opt(err)
	}
	return err
}

// WithInternalMessage updates `InternalMessage` field.
func (e *HTTPError) WithInternalMessage(message string) *HTTPError {
	e.InternalMessage = message
	return e
}

// WithInternalError updates `InternalError` field.
func (e *HTTPError) WithInternalError(err error) *HTTPError {
	e.InternalError = err
	return e
}

// WithErrorID updates `ErrorID` field.
func (e *HTTPError) WithErrorID(id string) *HTTPError {
	e.ErrorID = id
	return e
}

// WithShowError updates `ShowError` field.
func (e *HTTPError) WithShowError(flag bool) *HTTPError {
	e.ShowError = flag
	return e
}

// MarshalJSON implemenets `json.Marshal`.
func (e *HTTPError) MarshalJSON() ([]byte, error) {
	data := &httpError{
		Code:    e.Code,
		Message: e.Message,
		ErrorID: e.ErrorID,
	}
	if e.ShowError {
		if e.InternalError != nil {
			data.InternalError = e.InternalError.Error()
		}
		data.InternalMessage = e.InternalMessage
	}
	return json.Marshal(data)
}

// UnmarshalJSON implements `json.Unmarshal`.
func (e *HTTPError) UnmarshalJSON(b []byte) error {
	var data httpError
	err := json.Unmarshal(b, &data)
	if err != nil {
		return err
	}
	e.Code = data.Code
	e.Message = data.Message
	e.ErrorID = data.ErrorID
	if data.InternalError != "" {
		e.ShowError = true
		e.InternalError = errors.New(data.InternalError)
	}
	e.InternalMessage = data.InternalMessage
	return nil
}
