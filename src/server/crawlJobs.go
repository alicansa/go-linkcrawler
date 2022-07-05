package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type CrawlJobsService interface {
	AddJob(request CrawlJobRequest) CrawlJob
}

type CrawlJobStatus int

const (
	InProgress CrawlJobStatus = iota
	Completed
)

type CrawlJobRequest struct {
	BaseUrl string `json:"baseUrl"`
}

type CrawlJob struct {
	BaseUrl     string         `json:"baseUrl"`
	LastUpdated string         `json:"lastUpdated"`
	Status      CrawlJobStatus `json:"status"`
}

func (s *Server) registerCrawlJobsHandler(r *mux.Router) {
	r.HandleFunc("/crawlJobs", s.handleAddCrawlJob).Methods("POST")
}

func (s *Server) handleAddCrawlJob(rw http.ResponseWriter, r *http.Request) {
	// job, err := DecodeJSON[CrawlJobRequest](r.Body)
	var job CrawlJobRequest
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil || job.BaseUrl == "" {
		http.Error(rw, "Invalid json body", http.StatusBadRequest)
		return
	}

	log.Printf(job.BaseUrl)
}
