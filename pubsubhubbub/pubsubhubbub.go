// Package pubsubhubbub implements PubSubHubbub, as defined in
// http://pubsubhubbub.github.io/PubSubHubbub/pubsubhubbub-core-0.4.html.
package pubsubhubbub

import (
	"io"
)

const (
	// RelHub is the hub relation.
	RelHub = "hub"
	// RelUpdatesFrom is the updates-from relation.
	RelUpdatesFrom = "http://schemas.google.com/g/2010#updates-from"
)

// An Event is a notification sent by a publisher and received by a subscriber.
type Event interface {
	// MediaType returns the event's media type.
	MediaType() string
	// Topic returns the event's topic URL.
	Topic() string
	// WriteTo writes the event's body to w.
	WriteTo(w io.Writer) error
}

// ReadEventFunc reads an event.
type ReadEventFunc func(mediaType string, body io.Reader) (Event, error)
