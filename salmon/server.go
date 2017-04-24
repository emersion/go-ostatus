package salmon

import (
	"bytes"
	"encoding/xml"
	"net/http"

	"github.com/emersion/go-ostatus/activitystream"
)

type Backend interface {
	Reply(*activitystream.Entry) error
}

type handler struct {
	be Backend
}

func (h *handler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	if req.Method != "POST" {
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

	if err := h.be.Reply(entry); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
}

func NewHandler(be Backend) http.Handler {
	return &handler{be}
}
