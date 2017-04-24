package xrd

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
)

type HTTPError int

func (err HTTPError) Error() string {
	return "xrd: HTTP request failed"
}

func Get(url string) (*Resource, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, HTTPError(resp.StatusCode)
	}

	resource := new(Resource)
	switch resp.Header.Get("Content-Type") {
	case "application/xrd+xml", "application/xml", "text/xml":
		err = xml.NewDecoder(resp.Body).Decode(resource)
	case "application/jrd+json", "application/json", "":
		err = json.NewDecoder(resp.Body).Decode(resource)
	default:
		err = errors.New("xrd: unsupported format")
	}
	return resource, err
}
