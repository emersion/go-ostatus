package webfinger

import (
	"net/url"

	"github.com/emersion/go-ostatus/xrd"
)

func Get(domain, resourceURI string) (*xrd.Resource, error) {
	v := url.Values{}
	v.Set("resource", resourceURI)
	u := "https://" + domain + WellKnownPath + "?" + v.Encode()

	return xrd.Get(u)
}
