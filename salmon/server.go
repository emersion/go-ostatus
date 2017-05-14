package salmon

import (
	"bytes"
	"encoding/xml"
	"net/http"

	"github.com/emersion/go-ostatus/activitystream"
)

// A Backend is used to build salmon endpoints.
type Backend interface {
	// Notify is called when a salmon is pushed to the endpoint.
	Notify(*activitystream.Entry) error
}

type handler struct {
	be Backend
}

// ServeHTTP implements http.Handler.
func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != http.MethodPost {
		http.Error(resp, "Unsupported method", http.StatusBadRequest)
		return
	}

	switch req.Header.Get("Content-Type") {
	case "application/magic-envelope+xml", "application/xml":
		// Nothing to do
	default:
		http.Error(resp, "Unsupported content type", http.StatusBadRequest)
		return
	}

	env := new(MagicEnv)
	if err := xml.NewDecoder(req.Body).Decode(env); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO: get user public key
	// TODO: check signature
	b, err := env.UnverifiedData()
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	if env.Data.Type != "application/atom+xml" {
		http.Error(resp, "Unsupported content type within magic envelope", http.StatusBadRequest)
		return
	}

	entry := new(activitystream.Entry)
	if err := xml.NewDecoder(bytes.NewReader(b)).Decode(entry); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.be.Notify(entry); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

// NewHandler creates a new salmon endpoint.
func NewHandler(be Backend) http.Handler {
	return &handler{be}
}
