package salmon

import (
	"bytes"
	"encoding/xml"
	"encoding/json"
	"net/http"

	"github.com/emersion/go-ostatus/activitystream"
)

// A Backend is used to build salmon endpoints.
type Backend interface {
	PublicKeyBackend

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
		http.Error(resp, "Unsupported method", http.StatusMethodNotAllowed)
		return
	}

	env := new(MagicEnv)
	var err error
	switch req.Header.Get("Content-Type") {
	case "application/magic-envelope+xml", "application/xml":
		err = xml.NewDecoder(req.Body).Decode(env)
	case "application/magic-envelope+json", "application/json":
		err = json.NewDecoder(req.Body).Decode(env)
	default:
		http.Error(resp, "Unsupported content type", http.StatusBadRequest)
		return
	}
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

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

	accountURI := ""
	if entry.Author != nil {
		accountURI = entry.Author.AccountURI()
	}
	if accountURI == "" {
		http.Error(resp, "Cannot find account URI from payload", http.StatusBadRequest)
		return
	}

	pub, err := h.be.PublicKey(accountURI)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	if err := env.Verify(pub); err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.be.Notify(entry); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusAccepted)
}

// NewHandler creates a new salmon endpoint.
func NewHandler(be Backend) http.Handler {
	return &handler{be}
}
