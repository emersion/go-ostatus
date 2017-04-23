package pubsubhubbub

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/emersion/go-ostatus/activitystream"
)

type dummyBackend struct {
	topics map[string]chan<- *activitystream.Feed
}

func newDummyBackend() *dummyBackend {
	return &dummyBackend{
		topics: make(map[string]chan<- *activitystream.Feed),
	}
}

func (be *dummyBackend) Subscribe(topic string, notifies chan<- *activitystream.Feed) error {
	be.topics[topic] = notifies
	return nil
}

func (be *dummyBackend) Unsubscribe(notifies chan<- *activitystream.Feed) error {
	for topic, ch := range be.topics {
		if notifies == ch {
			delete(be.topics, topic)
			break
		}
	}

	return nil
}

type emptyReadCloser struct {}

func (r *emptyReadCloser) Read(b []byte) (int, error) {
	return 0, io.EOF
}

func (r *emptyReadCloser) Close() error {
	return nil
}

type roundTripper struct {
	h http.Handler
}

func (rt *roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Body == nil {
		req.Body = new(emptyReadCloser)
	}

	w := httptest.NewRecorder()
	rt.h.ServeHTTP(w, req)
	return w.Result(), nil
}

func Test(t *testing.T) {
	subscriberURL := "http://localhost/subscriber"
	publisherURL := "http://localhost/publisher"
	hubURL := publisherURL + "/hub"
	topicURL := publisherURL + "/topic.atom"

	be := newDummyBackend()
	pub := NewPublisher(be)
	sub := NewSubscriber(subscriberURL+"/webhook")
	pub.c.Transport = &roundTripper{sub}
	sub.c.Transport = &roundTripper{pub}

	notifies := make(chan *activitystream.Feed, 1)
	if err := sub.Subscribe(hubURL, topicURL, notifies); err != nil {
		t.Fatal("Expected no error when subscribing, got:", err)
	}

	updated := time.Now()
	sent := &activitystream.Feed{
		ID: topicURL,
		Title: "Test notification",
		Subtitle: "This is just a little test.",
		Updated: activitystream.NewTime(updated),
		Link: []activitystream.Link{
			{Rel: "self", Type: "application/atom+xml", Href: topicURL},
			{Rel: "hub", Href: hubURL},
		},
		Author: &activitystream.Person{
			ID: topicURL,
			Name: "Test subject #42",
			ObjectType: activitystream.ObjectPerson,
		},
		Entry: []*activitystream.Entry{
			{
				ID: "tag:localhost,2017-04-23:objectId=3865264:objectType=Status",
				Title: "My first post ever",
				Published: activitystream.NewTime(updated),
				Updated: activitystream.NewTime(updated),
				Content: &activitystream.Text{
					Type: "text/html",
					Body: "Hello World!",
				},
				ObjectType: activitystream.ObjectNote,
				Verb: activitystream.VerbPost,
			},
		},
	}

	// Send notification
	be.topics[topicURL] <- sent

	// Receive notification
	received := <-notifies

	sent.XMLName = received.XMLName
	if !reflect.DeepEqual(sent, received) {
		t.Error("Invalid notification, expected \n%+v\n but got \n%+v", sent, received)
	}
}
