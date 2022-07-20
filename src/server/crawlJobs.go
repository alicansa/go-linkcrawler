package server

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/alicansa/go-linkcrawler/crawler"
	"github.com/alicansa/go-linkcrawler/dal"
	"github.com/gorilla/mux"
)

type CrawlJobRequest struct {
	BaseUrl string `json:"baseUrl"`
}

type CrawlJob struct {
	BaseUrl     string             `json:"baseUrl"`
	LastUpdated string             `json:"lastUpdated"`
	Status      dal.CrawlJobStatus `json:"status"`
}

type CrawlJobsHandler struct {
	crawlJobRepository dal.CrawlJobRepository
	linkRepository     dal.LinkRepository
	newCrawler         func(pe crawler.CrawlPolicyExecuter) crawler.WebCrawler
}

func NewCrawlJobHandler(
	cjr dal.CrawlJobRepository,
	lr dal.LinkRepository,
	ncf func(pe crawler.CrawlPolicyExecuter) crawler.WebCrawler) *CrawlJobsHandler {
	return &CrawlJobsHandler{
		crawlJobRepository: cjr,
		linkRepository:     lr,
		newCrawler:         ncf,
	}
}

func (h *CrawlJobsHandler) registerCrawlJobsHandler(r *mux.Router) {
	r.HandleFunc("/crawlJobs/{id:[0-9]+}", h.getCrawlJob).Methods("GET")
	r.HandleFunc("/crawlJobs", h.getCrawlJobs).Methods("GET")
	r.HandleFunc("/crawlJobs", h.addCrawlJob).Methods("POST")
}

func (h *CrawlJobsHandler) getCrawlJob(rw http.ResponseWriter, r *http.Request) {

	jobId, err := strconv.Atoi(mux.Vars(r)["id"])

	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	job, err := h.crawlJobRepository.GetCrawlJob(jobId)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if job == (dal.CrawlJob{}) {
		http.Error(rw, "", http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(rw).Encode(job); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h *CrawlJobsHandler) getCrawlJobs(rw http.ResponseWriter, r *http.Request) {
	jobs, err := h.crawlJobRepository.GetCrawlJobs()

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-type", "application/json")
	if err := json.NewEncoder(rw).Encode(jobs); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}

func (h *CrawlJobsHandler) addCrawlJob(rw http.ResponseWriter, r *http.Request) {

	var job CrawlJobRequest
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil || job.BaseUrl == "" {
		http.Error(rw, "Invalid json body", http.StatusBadRequest)
		return
	}

	// check if job base url exists
	// if so return the job id
	existingJob, err := h.crawlJobRepository.GetCrawlJobForUrl(job.BaseUrl)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if existingJob != (dal.CrawlJob{}) {
		if err := json.NewEncoder(rw).Encode(existingJob.JobId); err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// add job with in progress status
	jobId, err := h.crawlJobRepository.AddCrawlJob(job.BaseUrl)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(rw).Encode(jobId); err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}

	pe := crawler.NewPolicyExecutor(
		"//a[@href[not(contains(.,'http')) and not(contains(.,'mailto:')) and not(contains(.,'tel:'))]]")
	c := h.newCrawler(pe)

	onLinksDiscovered := func(links []string) error {
		// add links to the db
		for _, link := range links {
			_, err := h.linkRepository.AddLink(link, jobId)

			if err != nil {
				return err
			}
		}

		return nil
	}

	go func() {
		var wg sync.WaitGroup
		wg.Add(1)

		go func() {
			defer wg.Done()
			//crawl
			_, err := c.Crawl(job.BaseUrl, onLinksDiscovered)
			if err != nil {
				log.Println(err.Error())
			}
		}()

		wg.Wait()
		// once crawl finished then update the job status
		h.crawlJobRepository.UpdateCrawlJobStatus(jobId, dal.Completed)
	}()
}
