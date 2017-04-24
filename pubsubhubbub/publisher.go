package pubsubhubbub

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/emersion/go-ostatus/activitystream"

	"log"
)

var DefaultLease = 24 * time.Hour

type Backend interface {
	Subscribe(topic string, notifies chan<- *activitystream.Feed) error
	Unsubscribe(notifies chan<- *activitystream.Feed) error
}

type pubSubscription struct {
	notifies  <-chan *activitystream.Feed
	callbacks map[string]time.Time
}

func (s *pubSubscription) receive(c *http.Client) error {
	// TODO: cancel subscription if lease expires

	for notif := range s.notifies {
		var b bytes.Buffer
		if err := notif.WriteTo(&b); err != nil {
			return err
		}

		for callback, _ := range s.callbacks {
			r := bytes.NewReader(b.Bytes())
			resp, err := c.Post(callback, "application/atom+xml", r)
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
	be            Backend
	c             *http.Client
	subscriptions map[string]*pubSubscription
}

func NewPublisher(be Backend) *Publisher {
	return &Publisher{
		be:            be,
		c:             new(http.Client),
		subscriptions: make(map[string]*pubSubscription),
	}
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

				s = &pubSubscription{
					notifies:  notifies,
					callbacks: make(map[string]time.Time),
				}
				go s.receive(p.c)

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
		challenge, err := generateChallenge()
		if err != nil {
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
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

		subResp, err := p.c.Get(u.String())
		if err != nil {
			log.Println("pubsubhubbub: cannot send HTTP request:", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if subResp.StatusCode/100 != 2 {
			log.Println("pubsubhubbub: HTTP request error:", subResp.Status)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		buf := make([]byte, len(challenge))
		if _, err := io.ReadFull(subResp.Body, buf); err != nil {
			log.Println("pubsubhubbub: cannot read HTTP response:", err)
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		} else if !bytes.Equal(buf, []byte(challenge)) {
			log.Println("pubsubhubbub: invalid challenge")
			http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		switch mode {
		case "subscribe":
			s.callbacks[callback] = lease
		case "unsubscribe":
			delete(s.callbacks, callback)
		}

		resp.WriteHeader(http.StatusAccepted)
	}
}
