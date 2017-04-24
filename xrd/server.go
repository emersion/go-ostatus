package xrd

import (
	"errors"
	"encoding/json"
	"encoding/xml"
	"net/http"
)

var ErrNoSuchResource = errors.New("xrd: no such resource")

type Backend interface {
	Resource(req *http.Request) (*Resource, error)
}

type handler struct {
	be Backend
}

func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
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
		err = xml.NewEncoder(resp).Encode(resource)
	}
	if err != nil {
		panic(err)
	}
}

func NewHandler(be Backend) http.Handler {
	return &handler{be}
}
