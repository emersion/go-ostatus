package xrd

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"mime"
	"net/http"
)

// An HTTPError is returned when an HTTP error has occured. Its value is the
// HTTP status code.
type HTTPError int

// Error implements error.
func (err HTTPError) Error() string {
	return "xrd: HTTP request failed"
}

// Get queries a resource.
func Get(url string) (*Resource, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, HTTPError(resp.StatusCode)
	}

	contentType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, err
	}

	resource := new(Resource)
	switch contentType {
	case "application/xrd+xml", "application/xml", "text/xml":
		err = xml.NewDecoder(resp.Body).Decode(resource)
	case "application/jrd+json", "application/json", "":
		err = json.NewDecoder(resp.Body).Decode(resource)
	default:
		err = errors.New("xrd: unsupported format: " + resp.Header.Get("Content-Type"))
	}
	return resource, err
}
