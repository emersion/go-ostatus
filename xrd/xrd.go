// Package xrd implements Extensible Resource Descriptor as defined in
// http://docs.oasis-open.org/xri/xrd/v1.0/xrd-1.0.html.
package xrd

import (
	"encoding/xml"
)

// A Resource is a resource descriptor.
type Resource struct {
	XMLName    xml.Name          `xml:"http://docs.oasis-open.org/ns/xri/xrd-1.0 XRD" json:"-"`
	Subject    string            `xml:"Subject" json:"subject,omitempty"`
	Aliases    []string          `xml:"Alias" json:"aliases,omitempty"`
	Properties map[string]string `xml:"-" json:"properties,omitempty"`
	Links      []*Link           `xml:"Link" json:"links,omitempty"`
}

// A Link provides a relationship between a resource and a URL.
type Link struct {
	Rel        string            `xml:"rel,attr,omitempty" json:"rel"`
	Type       string            `xml:"type,attr,omitempty" json:"type,omitempty"`
	Href       string            `xml:"href,attr,omitempty" json:"href,omitempty"`
	Template   string            `xml:"template,attr,omitempty" json:"template,omitempty"`
	Titles     map[string]string `xml:"-" json:"titles,omitempty"`
	Properties map[string]string `xml:"-" json:"properties,omitempty"`
}
