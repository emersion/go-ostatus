package hostmeta

import (
	"net/http"

	"github.com/emersion/go-ostatus/xrd"
)

type backend struct {
	resource *xrd.Resource
}

func (be *backend) Resource(*http.Request) (*xrd.Resource, error) {
	return be.resource, nil
}

func NewHandler(resource *xrd.Resource) http.Handler {
	return xrd.NewHandler(&backend{resource})
}
