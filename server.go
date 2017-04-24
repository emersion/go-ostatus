package ostatus

import (
	"net/http"

	"github.com/emersion/go-ostatus/xrd"
	"github.com/emersion/go-ostatus/xrd/hostmeta"
	"github.com/emersion/go-ostatus/xrd/webfinger"
	"github.com/emersion/go-ostatus/pubsubhubbub"
)

var HubPath = "/hub"

type handler struct {
	http.Handler
	be Backend
}

func NewHandler(be Backend, rootURL string) http.Handler {
	mux := http.NewServeMux()

	hostmetaResource := &xrd.Resource{
		Links: []*xrd.Link{
			{Rel: "lrdd", Type: "application/jrd+json", Template: rootURL+webfinger.WellKnownPathTemplate},
		},
	}

	mux.Handle(hostmeta.WellKnownPath, hostmeta.NewHandler(hostmetaResource))
	mux.Handle(webfinger.WellKnownPath, webfinger.NewHandler(be))
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
			http.Error(resp, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	return &handler{
		Handler: mux,
		be: be,
	}
}
