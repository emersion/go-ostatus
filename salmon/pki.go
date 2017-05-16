package salmon

import (
	"crypto"
	"errors"

	"github.com/emersion/go-ostatus/xrd"
	"github.com/emersion/go-ostatus/xrd/lrdd"
)

// ResourcePublicKey returns a resource's public key.
func ResourcePublicKey(resource *xrd.Resource) (crypto.PublicKey, error) {
	// TODO: multiple keys support
	var link *xrd.Link
	for _, l := range resource.Links {
		if l.Rel == RelMagicPublicKey {
			link = l
			break
		}
	}
	if link == nil {
		return nil, errors.New("salmon: missing magic-public-key link")
	}

	return ParsePublicKeyDataURL(link.Href)
}

// PublicKeyBackend represent a Public Key Infrastructure.
type PublicKeyBackend interface {
	// PublicKey retrieves a public key.
	PublicKey(accountURI string) (crypto.PublicKey, error)
}

type publicKeyBackend struct{}

// NewPublicKeyBackend returns a basic PublicKeyBackend that queries public keys
// with LRDD.
func NewPublicKeyBackend() PublicKeyBackend {
	return new(publicKeyBackend)
}

func (be *publicKeyBackend) PublicKey(accountURI string) (crypto.PublicKey, error) {
	resource, err := lrdd.Get(accountURI)
	// TODO: if err == lrdd.ErrNoHost, directly fetch accountURI (see section 8.2.3)
	if err != nil {
		return nil, err
	}

	return ResourcePublicKey(resource)
}
