package ostatus

import (
	"github.com/emersion/go-ostatus/activitystream"
	"github.com/emersion/go-ostatus/pubsubhubbub"
	"github.com/emersion/go-ostatus/salmon"
	"github.com/emersion/go-ostatus/xrd/webfinger"
)

// A Backend is used to create OStatus instances.
type Backend interface {
	webfinger.Backend
	pubsubhubbub.Backend
	salmon.Backend
	Feed(topic string) (*activitystream.Feed, error)
}
