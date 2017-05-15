package lrdd

import (
	"errors"
	"net/url"
	"strings"

	"github.com/emersion/go-ostatus/xrd"
	"github.com/emersion/go-ostatus/xrd/hostmeta"
)

var ErrNoHost = errors.New("lrdd: cannot extract host from URI's opaque data")

func executeTemplate(template, resourceURI string) string {
	return strings.Replace(template, "{uri}", url.QueryEscape(resourceURI), -1)
}

// Get retrieves a resource descriptor.
func Get(resourceURI string) (*xrd.Resource, error) {
	u, err := url.Parse(resourceURI)
	if err != nil {
		return nil, err
	}

	var host string
	if u.Host != "" {
		host = u.Host
	} else {
		parts := strings.SplitN(u.Opaque, "@", 2)
		if len(parts) != 2 {
			return nil, ErrNoHost
		}
		host = parts[1]
	}

	resource, err := hostmeta.Get(host)
	if err != nil {
		return nil, err
	}

	var link *xrd.Link
	for _, l := range resource.Links {
		if l.Rel == Rel {
			link = l
			break
		}
	}
	if link == nil {
		return nil, errors.New("lrdd: no lrdd link found in host-meta")
	}

	resourceURL := executeTemplate(link.Template, resourceURI)
	return xrd.Get(resourceURL)
}
