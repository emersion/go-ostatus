// Package webfinger implements WebFinger, as defined in
// https://tools.ietf.org/html/rfc7033.
package webfinger

const (
	WellKnownName         = "webfinger"
	WellKnownPath         = "/.well-known/webfinger"
	WellKnownPathTemplate = "/.well-known/webfinger?resource={uri}"
)
