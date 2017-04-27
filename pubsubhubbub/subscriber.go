package pubsubhubbub

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-ostatus/activitystream"

	"log"
)

type HTTPError struct {
	Status     string
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
	callbackURL  string
	lease        time.Time
	secret       string
	notifies     chan<- *activitystream.Feed
	subscribes   chan error
	unsubscribes chan error
}

type Subscriber struct {
	c             *http.Client
	callbackURL   string
	subscriptions map[string]*subscription
}

func NewSubscriber(callbackURL string) *Subscriber {
	return &Subscriber{
		c:             new(http.Client),
		callbackURL:   callbackURL,
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

	secret, err := generateChallenge()
	if err != nil {
		return err
	}

	u, err := url.Parse(s.callbackURL)
	if err != nil {
		return err
	}
	q := u.Query()
	q.Set("topic", topic)
	u.RawQuery = q.Encode()
	callbackURL := u.String()

	sub := &subscription{
		callbackURL:  callbackURL,
		notifies:     notifies,
		secret:       secret,
		subscribes:   make(chan error, 1),
		unsubscribes: make(chan error, 1),
	}
	s.subscriptions[topic] = sub

	data := make(url.Values)
	data.Set("hub.callback", callbackURL)
	data.Set("hub.mode", "subscribe")
	data.Set("hub.topic", topic)
	data.Set("hub.secret", secret)
	// hub.lease_seconds
	if err := s.request(hub, data); err != nil {
		return err
	}

	return <-sub.subscribes
}

func (s *Subscriber) Unsubscribe(hub, topic string) error {
	sub, ok := s.subscriptions[topic]
	if !ok {
		return errors.New("pubsubhubbub: no such subsciption")
	}

	data := make(url.Values)
	data.Set("hub.callback", sub.callbackURL)
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
		topic := query.Get("topic")

		sub, ok := s.subscriptions[topic]
		if !ok {
			http.Error(resp, "Invalid topic", http.StatusNotFound)
			return
		}

		var r io.Reader = req.Body
		var h hash.Hash
		if sub.secret != "" {
			h = hmac.New(sha1.New, []byte(sub.secret))
			r = io.TeeReader(r, h)
		}

		eventTopic, notifs, err := parseEvent(req.Header.Get("Content-Type"), r)
		if err != nil {
			http.Error(resp, "Invalid request body", http.StatusBadRequest)
			return
		}

		if eventTopic != topic {
			http.Error(resp, "Invalid topic", http.StatusNotFound)
			return
		}

		// Make sure the whole body has been read
		io.Copy(ioutil.Discard, r)

		// Check signature
		if h != nil {
			s := strings.TrimPrefix(req.Header.Get("X-Hub-Signature"), "sha1=")
			mac, err := hex.DecodeString(s)
			if err != nil || !hmac.Equal(mac, h.Sum(nil)) {
				// Invalid signature
				// Ignore message, do not return an error
				return
			}
		}

		sub.notifies <- notifs
	}
}
