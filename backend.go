package ostatus

import (
	"github.com/emersion/go-webfinger"
	"github.com/emersion/go-ostatus/activitystream"
	"github.com/emersion/go-ostatus/pubsubhubbub"
)

type Backend interface {
	webfinger.Backend
	pubsubhubbub.Backend
	Feed(topic string) (*activitystream.Feed, error)
}
