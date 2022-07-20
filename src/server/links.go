package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alicansa/go-linkcrawler/dal"
	"github.com/gorilla/mux"
)

type LinksHandler struct {
	linkRepository dal.LinkRepository
}

func NewLinksHandler(lr dal.LinkRepository) *LinksHandler {
	return &LinksHandler{
		linkRepository: lr,
	}
}

func (h *LinksHandler) registerLinksHandler(r *mux.Router) {
	r.HandleFunc("/links", h.getLinks)
}

func (h *LinksHandler) getLinks(rw http.ResponseWriter, req *http.Request) {

	// get crawl job id from the request query params
	crawlJobIdStr := req.URL.Query().Get("crawlJobId")

	crawlJobId, err := strconv.Atoi(crawlJobIdStr)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	links, err := h.linkRepository.GetLinks(crawlJobId)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(rw).Encode(links); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
	}
}
