package httputil

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Get performs HTTP GET request
func Get(url string) ([]byte, *HTTPError) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, NewHTTPError(err)
	}
	return doRequest(req)
}

// Put performs HTTP PUT request
func Put(url string, body []byte) ([]byte, *HTTPError) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, NewHTTPError(err)
	}
	return doRequest(req)
}

// Post performs HTTP POST request
func Post(url string, body []byte) ([]byte, *HTTPError) {
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, NewHTTPError(err)
	}
	return doRequest(req)
}

func doRequest(req *http.Request) ([]byte, *HTTPError) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, NewHTTPError(err)
	}
	return parseResponse(req, resp)
}

func parseResponse(req *http.Request, resp *http.Response) ([]byte, *HTTPError) {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, NewHTTPErrorFromReqRes(req, resp)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, NewHTTPErrorFromString(fmt.Sprintf("cannot read HTTP response body: %v", err), 0)
	}
	return data, nil
}
