package pubsubhubbub

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
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
	callbacks map[string]*pubCallback
}

type pubCallback struct {
	lease  time.Time
	secret string
}

func (s *pubSubscription) receive(c *http.Client) error {
	// TODO: cancel subscription if lease expires

	for notif := range s.notifies {
		var b bytes.Buffer
		if err := notif.WriteTo(&b); err != nil {
			return err
		}

		// TODO: retry if a request fails
		for callbackURL, cb := range s.callbacks {
			body := bytes.NewReader(b.Bytes())
			req, err := http.NewRequest("POST", callbackURL, body)
			if err != nil {
				log.Println("pubsubhubbub: failed create notification:", err)
				continue
			}

			req.Header.Set("Content-Type", "application/atom+xml")

			if cb.secret != "" {
				h := hmac.New(sha1.New, []byte(cb.secret))
				h.Write(b.Bytes())
				sig := hex.EncodeToString(h.Sum(nil))
				req.Header.Set("X-Hub-Signature", "sha1="+sig)
			}

			resp, err := c.Do(req)
			if err != nil {
				log.Println("pubsubhubbub: failed to push notification:", err)
				continue
			}
			resp.Body.Close()

			if resp.StatusCode/100 != 2 {
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

	if req.Method != "POST" {
		http.Error(resp, "Invalid method", http.StatusBadRequest)
		return
	}

	mode := req.FormValue("hub.mode")
	callbackURL := req.FormValue("hub.callback")
	topicURL := req.FormValue("hub.topic")
	secret := req.FormValue("hub.secret")

	if mode != "subscribe" && mode != "unsubscribe" {
		http.Error(resp, "Invalid mode", http.StatusBadRequest)
		return
	}
	if len(secret) > 200 {
		http.Error(resp, "Secret too long", http.StatusBadRequest)
		return
	}

	u, err := url.ParseRequestURI(callbackURL)
	if err != nil {
		http.Error(resp, "Invalid callback URL", http.StatusBadRequest)
		return
	}
	q := u.Query()
	q.Set("hub.topic", topicURL)

	// TODO: do this in another goroutine

	// Subscribe if necessary
	var lease time.Time
	s, ok := p.subscriptions[topicURL]
	switch mode {
	case "subscribe":
		if !ok {
			notifies := make(chan *activitystream.Feed)
			if err := p.be.Subscribe(topicURL, notifies); err != nil {
				if deniedErr, ok := err.(DeniedError); ok {
					// Send denied notification
					q.Set("hub.mode", "denied")
					q.Set("hub.reason", string(deniedErr))
					u.RawQuery = q.Encode()

					subResp, err := p.c.Get(u.String())
					if err != nil {
						log.Println("pubsubhubbub: cannot send HTTP request:", err)
						return
					}
					subResp.Body.Close()
					return
				} else {
					log.Printf("pubsubhubbub: backend returned error when subscribing to %q: %v\n", topicURL, err)
					http.Error(resp, "Internal Server Error", http.StatusInternalServerError)
					return
				}
			}

			s = &pubSubscription{
				notifies:  notifies,
				callbacks: make(map[string]*pubCallback),
			}
			go s.receive(p.c)

			p.subscriptions[topicURL] = s
		}

		lease = time.Now().Add(DefaultLease)
	case "unsubscribe":
		if !ok {
			return
		} else if _, ok := s.callbacks[callbackURL]; !ok {
			return
		}
	}

	// Verify
	challenge, err := generateChallenge()
	if err != nil {
		log.Println("pubsubhubbub: cannot generate challenge:", err)
		return
	}

	q.Set("hub.mode", mode)
	q.Set("hub.challenge", challenge)
	if mode == "subscribe" {
		q.Set("hub.lease_seconds", strconv.Itoa(int(lease.Sub(time.Now()).Seconds())))
	}
	u.RawQuery = q.Encode()

	subResp, err := p.c.Get(u.String())
	if err != nil {
		log.Println("pubsubhubbub: cannot send HTTP request:", err)
		return
	}
	defer subResp.Body.Close()

	if subResp.StatusCode/100 != 2 {
		log.Println("pubsubhubbub: HTTP request error:", subResp.Status)
		return
	}

	buf := make([]byte, len(challenge))
	if _, err := io.ReadFull(subResp.Body, buf); err != nil {
		log.Println("pubsubhubbub: cannot read HTTP response:", err)
		return
	} else if !bytes.Equal(buf, []byte(challenge)) {
		log.Println("pubsubhubbub: invalid challenge")
		return
	}

	switch mode {
	case "subscribe":
		s.callbacks[callbackURL] = &pubCallback{
			lease: lease,
			secret: secret,
		}
	case "unsubscribe":
		delete(s.callbacks, callbackURL)
	}

	resp.WriteHeader(http.StatusAccepted)
}
