package xrd

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
)

// ErrNoSuchResource can be returned by a Backend if a resource doesn't exist.
var ErrNoSuchResource = errors.New("xrd: no such resource")

// A Backend is used to build an XRD endpoint.
type Backend interface {
	Resource(req *http.Request) (*Resource, error)
}

type handler struct {
	be Backend
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// Default Access-Control-Allow-Origin to *
	if resp.Header().Get("Access-Control-Allow-Origin") == "" {
		resp.Header().Set("Access-Control-Allow-Origin", "*")
	}

	resource, err := h.be.Resource(req)
	if err == ErrNoSuchResource {
		http.NotFound(resp, req)
		return
	} else if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: properly parse Accept header
	switch req.Header.Get("Accept") {
	case "application/jrd+json", "application/json":
		resp.Header().Set("Content-Type", "application/jrd+json")
		err = json.NewEncoder(resp).Encode(resource)
	default:
		resp.Header().Set("Content-Type", "application/xrd+xml")
		if _, err = io.WriteString(resp, xml.Header); err != nil {
			break
		}
		err = xml.NewEncoder(resp).Encode(resource)
	}
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewHandler creates a new XRD endpoint.
func NewHandler(be Backend) http.Handler {
	return &handler{be}
}
