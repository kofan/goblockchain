package httputil

import (
	"fmt"
	"net/http"
)

// HTTPError represent the error occurred as a result of HTTP request attempt
type HTTPError struct {
	Message    string
	StatusCode int
}

func NewHTTPError(err error) *HTTPError {
	HTTPError := &HTTPError{err.Error(), 0}
	return HTTPError
}

func NewHTTPErrorFromString(message string, statusCode int) *HTTPError {
	HTTPError := &HTTPError{message, statusCode}
	return HTTPError
}

func NewHTTPErrorFromReqRes(req *http.Request, resp *http.Response) *HTTPError {
	message := fmt.Sprintf("HTTP request failed %s %s - %s", req.Method, req.URL, resp.Status)
	return NewHTTPErrorFromString(message, resp.StatusCode)
}

func (err *HTTPError) Error() string {
	return err.Message
}
