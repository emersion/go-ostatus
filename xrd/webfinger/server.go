package webfinger

import (
	"net/http"

	"github.com/emersion/go-ostatus/xrd"
)

type Backend interface {
	Resource(uri string, rel []string) (*xrd.Resource, error)
}

type backend struct {
	Backend
}

func (be *backend) Resource(req *http.Request) (*xrd.Resource, error) {
	q := req.URL.Query()
	resourceURI := q.Get("resource")
	// TODO: rel
	return be.Backend.Resource(resourceURI, nil)
}

func NewHandler(be Backend) http.Handler {
	return xrd.NewHandler(&backend{be})
}
