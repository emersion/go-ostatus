package ostatus

import (
	"github.com/emersion/go-ostatus/activitystream"
	"github.com/emersion/go-ostatus/pubsubhubbub"
	"github.com/emersion/go-ostatus/salmon"
	"github.com/emersion/go-ostatus/xrd/webfinger"
)

type Backend interface {
	webfinger.Backend
	pubsubhubbub.Backend
	salmon.Backend
	Feed(topic string) (*activitystream.Feed, error)
}
