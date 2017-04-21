// Package pubsubhubbub implements PubSubHubbub, as defined in
// http://pubsubhubbub.github.io/PubSubHubbub/pubsubhubbub-core-0.4.html.
package pubsubhubbub

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/emersion/go-ostatus/activitystream"

	"log"
)

var DefaultLease = 24 * time.Hour

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
	Unsubscribe(notifies chan<- *activitystream.Feed) error
}

type pubSubscription struct {
	notifies <-chan *activitystream.Feed
	callbacks map[string]time.Time
}

func (s *pubSubscription) receive() error {
	// TODO: cancel subscription if lease expires

	for notif := range s.notifies {
		var b bytes.Buffer
		if err := notif.WriteTo(&b); err != nil {
			return err
		}

		for callback, _ := range s.callbacks {
			r := bytes.NewReader(b.Bytes())
			resp, err := http.Post(callback, "application/atom+xml", r)
			if err != nil {
				// TODO: retry
				log.Println("pubsubhubbub: failed to push notification:", err)
				continue
			}

			resp.Body.Close()
			if resp.StatusCode/100 != 2 {
				// TODO: retry
				log.Println("pubsubhubbub: failed to push notification:", resp.StatusCode, resp.Status)
				continue
			}
		}
	}

	return nil
}

type Publisher struct {
	be Backend
	c *http.Client
	subscriptions map[string]*pubSubscription
}

func (p *Publisher) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method == "POST" {
		mode := req.FormValue("hub.mode")
		callback := req.FormValue("hub.callback")
		topic := req.FormValue("hub.topic")

		if mode != "subscribe" && mode != "unsubscribe" {
			http.Error(resp, "Bad Request", http.StatusBadRequest)
			return
		}

		// TODO: do this in another goroutine

		// Subscribe if necessary
		var lease time.Time
		s, ok := p.subscriptions[topic]
		switch mode {
		case "subscribe":
			if !ok {
				notifies := make(chan *activitystream.Feed)
				if err := p.be.Subscribe(topic, notifies); err != nil {
					if _, ok := err.(DeniedError); ok {
						// TODO: send denied notification
						return
					} else {
						log.Printf("pubsubhubbub: backend returned error when subscribing to %q: %v\n", topic, err)
						http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
						return
					}
				}

				s = &pubSubscription{notifies: notifies}
				go s.receive()

				p.subscriptions[topic] = s
			}

			lease = time.Now().Add(DefaultLease)
		case "unsubscribe":
			if !ok {
				return
			} else {
				// TODO: check that callback is in s.callbacks
			}
		}

		// Verify
		challenge := generateChallenge()
		u, err := url.ParseRequestURI(callback)
		if err != nil {
			http.Error(resp, "Bad Request", http.StatusBadRequest)
			return
		}
		q := u.Query()
		q.Set("hub.mode", mode)
		q.Set("hub.topic", topic)
		q.Set("hub.challenge", challenge)
		if mode == "subscribe" {
			q.Set("hub.lease_seconds", strconv.Itoa(int(lease.Sub(time.Now()).Seconds())))
		}
		u.RawQuery = q.Encode()

		resp, err := http.Get(u.String())
		if err != nil {
			return
		}

		buf := make([]byte, len(challenge))
		if _, err := io.ReadFull(resp.Body, buf); err != nil {
			return
		} else if !bytes.Equal(buf, []byte(challenge)) {
			return
		}

		switch mode {
		case "subscribe":
			s.callbacks[callback] = lease
		case "unsubscribe":
			delete(s.callbacks, callback)
		}
	}
}
