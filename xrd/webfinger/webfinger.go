// Package webfinger implements WebFinger, as defined in
// https://tools.ietf.org/html/rfc7033.
package webfinger

const (
	WellKnownName         = "webfinger"
	WellKnownPath         = "/.well-known/webfinger"
	WellKnownPathTemplate = "/.well-known/webfinger?resource={uri}"
)

// RelProfilePage is the profile-page relation.
const RelProfilePage = "http://webfinger.net/rel/profile-page"
