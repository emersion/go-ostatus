package hostmeta

import (
	"net/http"

	"github.com/emersion/go-ostatus/xrd"
)

type handler struct {
	resource *xrd.Resource
}

func (be *handler) Resource(*http.Request) (*xrd.Resource, error) {
	return be.resource, nil
}

func NewHandler(resource *xrd.Resource) http.Handler {
	return xrd.NewHandler(&handler{resource})
}
