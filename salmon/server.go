package salmon

import (
	"net/http"
)

type Backend interface{}

type handler struct{
	be Backend
}

func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// TODO
}

func NewHandler(be Backend) http.Handler {
	return &handler{be}
}
