// Package pubsubhubbub implements PubSubHubbub, as defined in
// http://pubsubhubbub.github.io/PubSubHubbub/pubsubhubbub-core-0.4.html.
package pubsubhubbub

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/emersion/go-ostatus/activitystream"

	"log"
)

type HTTPError struct {
	Status string
	StatusCode int
}

func (err *HTTPError) Error() string {
	return "pubsubhubbub: HTTP request failed: " + err.Status
}

type DeniedError string

func (err DeniedError) Error() string {
	return "pubsubhubbub: subscription denied: " + string(err)
}

func parseEvent(mediaType string, body io.Reader) (topic string, feed *activitystream.Feed, err error) {
	if mediaType != "application/atom+xml" {
		err = errors.New("pubsubhubbub: unsupported notification media type")
		return
	}

	feed, err = activitystream.Read(body)
	if err != nil {
		return
	}

	// Find topic
	for _, link := range feed.Link {
		if link.Rel == "self" {
			topic = link.Href
			break
		}
	}
	if topic == "" {
		err = errors.New("pubsubhubbub: no topic found in event")
		return
	}

	return
}

type subscription struct {
	lease time.Time
	notifies chan<- *activitystream.Feed
	subscribes chan error
	unsubscribes chan error
}

type Subscriber struct {
	c *http.Client
	callbackURL string
	subscriptions map[string]*subscription
}

func NewSubscriber(callbackURL string) *Subscriber {
	return &Subscriber{
		c: new(http.Client),
		callbackURL: callbackURL,
		subscriptions: make(map[string]*subscription),
	}
}

func (s *Subscriber) request(hub string, data url.Values) error {
	resp, err := s.c.PostForm(hub, data)
	if err != nil {
		return err
	}
	resp.Body.Close() // We don't need the response body

	if resp.StatusCode != http.StatusAccepted {
		return &HTTPError{resp.Status, resp.StatusCode}
	}

	return nil
}

func (s *Subscriber) Subscribe(hub, topic string, notifies chan<- *activitystream.Feed) error {
	if _, ok := s.subscriptions[topic]; ok {
		return errors.New("pubsubhubbub: already subscribed")
	}

	data := make(url.Values)
	data.Set("hub.callback", s.callbackURL)
	data.Set("hub.mode", "subscribe")
	data.Set("hub.topic", topic)
	// hub.lease_seconds, hub.secret
	if err := s.request(hub, data); err != nil {
		return err
	}

	sub := &subscription{
		notifies: notifies,
		subscribes: make(chan error, 1),
		unsubscribes: make(chan error, 1),
	}
	s.subscriptions[topic] = sub
	return <-sub.subscribes
}

func (s *Subscriber) Unsubscribe(hub, topic string) error {
	sub, ok := s.subscriptions[topic]
	if !ok {
		return errors.New("pubsubhubbub: no such subsciption")
	}

	data := make(url.Values)
	data.Set("hub.callback", s.callbackURL)
	data.Set("hub.mode", "unsubscribe")
	data.Set("hub.topic", topic)
	if err := s.request(hub, data); err != nil {
		return err
	}

	return <-sub.unsubscribes
}

func (s *Subscriber) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	query := req.URL.Query()
	if mode := query.Get("hub.mode"); mode != "" {
		topic := query.Get("hub.topic")

		sub, ok := s.subscriptions[topic]
		if !ok {
			http.Error(resp, "Not Found", http.StatusNotFound)
			return
		}

		switch mode {
		case "denied":
			reason := query.Get("hub.reason")
			log.Printf("pubsubhubbub: publisher denied request for topic %q (reason: %v)\n", topic, reason)
			delete(s.subscriptions, topic)
			close(sub.notifies)
			sub.subscribes <- DeniedError(reason)
			close(sub.subscribes)
			return
		case "subscribe":
			log.Printf("pubsubhubbub: publisher accepted subscription for topic %q\n", topic)
			lease, err := strconv.Atoi(query.Get("hub.lease_seconds"))
			if err != nil {
				http.Error(resp, "Bad Request", http.StatusBadRequest)
				return
			}
			sub.lease = time.Now().Add(time.Duration(lease) * time.Second)
			close(sub.subscribes)
		case "unsubscribe":
			log.Printf("pubsubhubbub: publisher accepted unsubscription for topic %q\n", topic)
			delete(s.subscriptions, topic)
			close(sub.notifies)
			close(sub.unsubscribes)
		default:
			http.Error(resp, "Bad Request", http.StatusBadRequest)
			return
		}

		resp.Write([]byte(query.Get("hub.challenge")))
	} else {
		topic, notifs, err := parseEvent(req.Header.Get("Content-Type"), req.Body)
		if err != nil {
			http.Error(resp, "Bad Request", http.StatusBadRequest)
			return
		}

		sub, ok := s.subscriptions[topic]
		if !ok {
			http.Error(resp, "Not Found", http.StatusNotFound)
			return
		}

		sub.notifies <- notifs
	}
}

type Backend interface {
	Subscribe(topic string, notifies chan<- *activitystream.Feed) error
}

type Publisher struct {
	c *http.Client
}

func (p *Publisher) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	// TODO
}
