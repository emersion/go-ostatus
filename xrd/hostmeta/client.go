package hostmeta

import (
	"github.com/emersion/go-ostatus/xrd"
)

func Get(domain string) (*xrd.Resource, error) {
	u := "https://" + domain + WellKnownPath
	return xrd.Get(u)
}
