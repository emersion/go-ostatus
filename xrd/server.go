package xrd

import (
	"errors"
	"encoding/json"
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

	// TODO: parse Accept header field
	resp.Header().Set("Content-Type", "application/jrd+json")
	if err := json.NewEncoder(resp).Encode(resource); err != nil {
		panic(err)
	}
}

func NewHandler(be Backend) http.Handler {
	return &handler{be}
}
