// Package activitystream implements Atom Activity Streams 1.0, as defined in
// http://activitystrea.ms/specs/atom/1.0/.
package activitystream

import (
	"encoding/xml"
	"io"
	"net/http"
	"time"
)

type Feed struct {
	XMLName  xml.Name `xml:"http://www.w3.org/2005/Atom feed"`
	ID       string   `xml:"id"`
	Title    string   `xml:"title"`
	Subtitle string   `xml:"subtitle,omitempty"`
	Updated  Time     `xml:"updated"`
	Logo     string   `xml:"logo,omitempty"`
	Author   *Person  `xml:"author"`
	Link     []Link   `xml:"link"`
	Entry    []*Entry `xml:"entry"`
}

func Read(r io.Reader) (*Feed, error) {
	feed := new(Feed)
	err := xml.NewDecoder(r).Decode(feed)
	return feed, err
}

type HTTPError struct {
	Status     string
	StatusCode int
}

func (err *HTTPError) Error() string {
	return "activitystream: HTTP request failed"
}

func Get(url string) (*Feed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &HTTPError{resp.Status, resp.StatusCode}
	}

	return Read(resp.Body)
}

func (feed *Feed) WriteTo(w io.Writer) error {
	return xml.NewEncoder(w).Encode(feed)
}

type Entry struct {
	ID        string  `xml:"id"`
	Title     string  `xml:"title"`
	Link      []Link  `xml:"link"`
	Published Time    `xml:"published"`
	Updated   Time    `xml:"updated"`
	Author    *Person `xml:"author"`
	Summary   *Text   `xml:"summary"`
	Content   *Text   `xml:"content"`

	ObjectType ObjectType `xml:"http://activitystrea.ms/spec/1.0/ object-type,omitempty"`
	Verb       Verb       `xml:"http://activitystrea.ms/spec/1.0/ verb,omitempty"`
	Object     *Entry     `xml:"http://activitystrea.ms/spec/1.0/ object"`

	InReplyTo *Link `xml:"http://purl.org/syndication/thread/1.0 in-reply-to"`
}

type Link struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`

	HrefLang string `xml:"hreflang,attr,omitempty"`
	Title    string `xml:"title,attr,omitempty"`
	Length   uint   `xml:"length,attr,omitempty"`

	ObjectType ObjectType `xml:"http://ostatus.org/schema/1.0 object-type,omitempty"`

	MediaWidth  uint `xml:"http://purl.org/syndication/atommedia width,omitempty"`
	MediaHeight uint `xml:"http://purl.org/syndication/atommedia height,omitempty"`
}

type Person struct {
	ID      string `xml:"id"`
	URI     string `xml:"uri,omitempty"`
	Name    string `xml:"name"`
	Email   string `xml:"email,omitempty"`
	Summary string `xml:"summary,omitempty"`
	Link    []Link `xml:"link"`

	ObjectType ObjectType `xml:"http://activitystrea.ms/spec/1.0/ object-type,omitempty"`

	PreferredUsername string `xml:"http://portablecontacts.net/spec/1.0 preferredUsername,omitempty"`
	DisplayName       string `xml:"http://portablecontacts.net/spec/1.0 displayName,omitempty"`
	Note              string `xml:"http://portablecontacts.net/spec/1.0 note,omitempty"`
}

type Text struct {
	Type string `xml:"type,attr"`
	Body string `xml:",chardata"`
}

const timeLayout = "2006-01-02T15:04:05-07:00"

type Time string

func NewTime(t time.Time) Time {
	return Time(t.Format(timeLayout))
}

func (t Time) Time() (time.Time, error) {
	return time.Parse(timeLayout, string(t))
}

type ObjectType string

const (
	ObjectPerson     ObjectType = "http://activitystrea.ms/schema/1.0/person"
	ObjectNote                  = "http://activitystrea.ms/schema/1.0/note"
	ObjectComment               = "http://activitystrea.ms/schema/1.0/comment"
	ObjectCollection            = "http://activitystrea.ms/schema/1.0/collection"
)

type Verb string

const (
	VerbPost  Verb = "http://activitystrea.ms/schema/1.0/post"
	VerbShare      = "http://activitystrea.ms/schema/1.0/share"
)
