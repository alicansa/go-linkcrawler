package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Link struct {
	Url string `json:"url"`
}

type LinksService interface {
	GetLinks() []Link
}

func (s *Server) registerLinksHandler(r *mux.Router) {
	r.HandleFunc("/links", s.getLinks)
}

func (s *Server) getLinks(rw http.ResponseWriter, req *http.Request) {
	rw.Header().Set("Content-type", "application/json")

	links := [2]Link{
		{Url: "test"},
		{Url: "test2"},
	}

	if err := json.NewEncoder(rw).Encode(links); err != nil {
		log.Printf(err.Error())
		return
	}
}
