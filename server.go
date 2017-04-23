package ostatus

import (
	"net/http"

	"github.com/emersion/go-webfinger"
	"github.com/emersion/go-ostatus/pubsubhubbub"
)

var HubPath = "/hub"

type handler struct {
	http.Handler
	be Backend
}

func NewHandler(be Backend) http.Handler {
	mux := http.NewServeMux()

	mux.Handle(webfinger.Path, webfinger.NewHandler(be))
	mux.Handle(HubPath, pubsubhubbub.NewPublisher(be))

	mux.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		topic := req.URL.String()
		feed, err := be.Feed(topic)
		if err != nil {
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}

		resp.Header().Set("Content-Type", "application/atom+xml")
		if err := feed.WriteTo(resp); err != nil {
			panic(err)
		}
	})

	return &handler{
		Handler: mux,
		be: be,
	}
}
