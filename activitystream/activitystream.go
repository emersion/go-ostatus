// Package activitystream implements Atom Activity Streams 1.0, as defined in
// http://activitystrea.ms/specs/atom/1.0/.
package activitystream

import (
	"encoding/xml"
	"io"
	"net/http"
	"time"
)

// A Feed is an activity stream feed.
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

// Read parses a feed from r.
func Read(r io.Reader) (*Feed, error) {
	feed := new(Feed)
	err := xml.NewDecoder(r).Decode(feed)
	return feed, err
}

// An HTTPError is an HTTP error. Its value is the HTTP status code.
type HTTPError int

// Error implements error.
func (err HTTPError) Error() string {
	return "activitystream: HTTP request failed"
}

// Get retrieves a feed located at a given URL.
func Get(url string) (*Feed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, HTTPError(resp.StatusCode)
	}

	return Read(resp.Body)
}

// WriteTo writes the feed to w.
func (feed *Feed) WriteTo(w io.Writer) error {
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	return xml.NewEncoder(w).Encode(feed)
}

// An Entry is a feed item.
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

	InReplyTo *InReplyTo `xml:"http://purl.org/syndication/thread/1.0 in-reply-to"`
}

// A Link provides a relationship between an entry or a person and a URL.
type Link struct {
	Rel  string `xml:"rel,attr,omitempty"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`

	HrefLang string `xml:"hreflang,attr,omitempty"`
	Title    string `xml:"title,attr,omitempty"`
	Length   uint   `xml:"length,attr,omitempty"`

	ObjectType ObjectType `xml:"http://ostatus.org/schema/1.0 object-type,attr,omitempty"`

	MediaWidth  uint `xml:"http://purl.org/syndication/atommedia width,attr,omitempty"`
	MediaHeight uint `xml:"http://purl.org/syndication/atommedia height,attr,omitempty"`
}

// A Person is a person.
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

// A Text has a type and body.
type Text struct {
	Type string `xml:"type,attr"`
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr,omitempty"`
	Body string `xml:",chardata"`
}

const timeLayout = "2006-01-02T15:04:05-07:00"

// A Time is a formatted time.
type Time string

// NewTime formats a time.
func NewTime(t time.Time) Time {
	return Time(t.Format(timeLayout))
}

// Time parses a formatted time.
func (t Time) Time() (time.Time, error) {
	return time.Parse(timeLayout, string(t))
}

// InReplyTo is used to indicate that an entry is a response to another
// resource.
type InReplyTo struct {
	Ref    string `xml:"ref,attr"`
	Href   string `xml:"href,attr,omitempty"`
	Source string `xml:"source,attr,omitempty"`
	Type   string `xml:"type,attr,omitempty"`
}

// An ObjectType describes the type of an object.
type ObjectType string

const (
	ObjectActivity   ObjectType = "http://activitystrea.ms/schema/1.0/activity"
	ObjectNote                  = "http://activitystrea.ms/schema/1.0/note"
	ObjectComment               = "http://activitystrea.ms/schema/1.0/comment"
	ObjectPerson                = "http://activitystrea.ms/schema/1.0/person"
	ObjectCollection            = "http://activitystrea.ms/schema/1.0/collection"
	ObjectGroup                 = "http://activitystrea.ms/schema/1.0/group"
)

// A Verb describes an action.
type Verb string

const (
	VerbPost          Verb = "http://activitystrea.ms/schema/1.0/post"
	VerbShare              = "http://activitystrea.ms/schema/1.0/share"
	VerbFavorite           = "http://activitystrea.ms/schema/1.0/favorite"
	VerbUnfavorite         = "http://activitystrea.ms/schema/1.0/unfavorite"
	VerbDelete             = "http://activitystrea.ms/schema/1.0/delete"
	VerbFollow             = "http://activitystrea.ms/schema/1.0/follow"
	VerbRequestFriend      = "http://activitystrea.ms/schema/1.0/request-friend"
	VerbAuthorize          = "http://activitystrea.ms/schema/1.0/authorize"
	VerbReject             = "http://activitystrea.ms/schema/1.0/reject"
	VerbUnfollow           = "http://ostatus.org/schema/1.0/unfollow"
	VerbBlock              = "http://mastodon.social/schema/1.0/block"
	VerbUnblock            = "http://mastodon.social/schema/1.0/unblock"
)

const (
	CollectionPublic = "http://activityschema.org/collection/public"
)
