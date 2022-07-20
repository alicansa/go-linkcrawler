package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicansa/go-linkcrawler/crawler"
	"github.com/alicansa/go-linkcrawler/dal"
	"github.com/alicansa/go-linkcrawler/mocks"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type CrawlJobsTest struct {
	server           *httptest.Server
	controller       *gomock.Controller
	mockCrawlJobRepo *mocks.MockCrawlJobRepository
	mockLinkRepo     *mocks.MockLinkRepository
	mockWebCrawler   *mocks.MockWebCrawler
}

func TestCrawlJobs(t *testing.T) {
	// setup + teardown
	cjt := &CrawlJobsTest{}
	tds := cjt.setupSuite(t)
	defer tds(t)

	t.Run("Test getCrawlJob returns internal server error on db issue", cjt.testGetCrawlJobReturnsInternalServerErrorOnDbError)
	t.Run("Test getCrawlJob returns not found if job doesn't exist", cjt.testGetCrawlJobReturnsNotFoundIfJobDoesntExist)
	t.Run("Test successful getCrawlJob call", cjt.testSuccessfulGetCrawlJob)
	t.Run("Test getCrawlJobs returns internal server error on db issue", cjt.testGetCrawlJobsReturnsInternalServerErrorOnDbError)
	t.Run("Test successful getCrawlJobs call", cjt.testSuccessfulGetCrawlJobs)
	t.Run("Test add crawlJobs returns bad request on invalid json", cjt.testAddCrawlJobReturnsBadRequestOnInvalidJson)
	t.Run("Test add crawlJobs returns job id if url already added", cjt.testAddCrawlJobReturnsJobIdIfAlreadyAdded)
	t.Run("Test successful add crawlJobs", cjt.testSuccessfulAddCrawlJob)
}

func (cjt *CrawlJobsTest) setupSuite(t *testing.T) func(t *testing.T) {

	//create mocks
	cjt.controller = gomock.NewController(t)
	mockLinkRepo := mocks.NewMockLinkRepository(cjt.controller)
	cjt.mockLinkRepo = mockLinkRepo
	mockCrawlJobRepo := mocks.NewMockCrawlJobRepository(cjt.controller)
	cjt.mockCrawlJobRepo = mockCrawlJobRepo
	mockWebCrawler := mocks.NewMockWebCrawler(cjt.controller)
	cjt.mockWebCrawler = mockWebCrawler

	//create router and link it up
	r := mux.NewRouter()
	crawlJobsHandler := NewCrawlJobHandler(
		mockCrawlJobRepo,
		mockLinkRepo,
		func(pe crawler.CrawlPolicyExecuter) crawler.WebCrawler {
			return mockWebCrawler
		},
	)

	crawlJobsHandler.registerCrawlJobsHandler(r)
	cjt.server = httptest.NewServer(r)

	return func(t *testing.T) {
		cjt.server.Close()
	}
}

func (cjt *CrawlJobsTest) testGetCrawlJobReturnsInternalServerErrorOnDbError(t *testing.T) {

	var emptyJob dal.CrawlJob
	cjt.mockCrawlJobRepo.EXPECT().GetCrawlJob(123).Return(emptyJob, errors.New("db error")).Times(1)

	resp, err := http.Get(cjt.server.URL + "/crawlJobs/123")

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func (cjt *CrawlJobsTest) testGetCrawlJobReturnsNotFoundIfJobDoesntExist(t *testing.T) {

	var emptyJob dal.CrawlJob
	cjt.mockCrawlJobRepo.EXPECT().GetCrawlJob(123).Return(emptyJob, nil).Times(1)

	resp, err := http.Get(cjt.server.URL + "/crawlJobs/123")

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func (cjt *CrawlJobsTest) testSuccessfulGetCrawlJob(t *testing.T) {

	job := dal.CrawlJob{
		BaseUrl:     "test",
		LastUpdated: "10:11:14",
		Status:      dal.InProgress,
		JobId:       123,
	}
	cjt.mockCrawlJobRepo.EXPECT().GetCrawlJob(123).Return(job, nil).Times(1)

	resp, err := http.Get(cjt.server.URL + "/crawlJobs/123")

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	//decode body
	var respJob dal.CrawlJob
	if err = json.NewDecoder(resp.Body).Decode(&respJob); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, job, respJob)
}

func (cjt *CrawlJobsTest) testGetCrawlJobsReturnsInternalServerErrorOnDbError(t *testing.T) {

	var emptyJobs []dal.CrawlJob
	cjt.mockCrawlJobRepo.EXPECT().GetCrawlJobs().Return(emptyJobs, errors.New("db error")).Times(1)

	resp, err := http.Get(cjt.server.URL + "/crawlJobs")

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func (cjt *CrawlJobsTest) testSuccessfulGetCrawlJobs(t *testing.T) {

	jobs := []dal.CrawlJob{
		{
			BaseUrl:     "test",
			LastUpdated: "10:11:10",
			Status:      dal.InProgress,
			JobId:       123,
		},
		{
			BaseUrl:     "test2",
			LastUpdated: "10:11:12",
			Status:      dal.Completed,
			JobId:       124,
		},
	}
	cjt.mockCrawlJobRepo.EXPECT().GetCrawlJobs().Return(jobs, nil).Times(1)

	resp, err := http.Get(cjt.server.URL + "/crawlJobs")

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	//decode body
	var respJob []dal.CrawlJob
	if err = json.NewDecoder(resp.Body).Decode(&respJob); err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, jobs, respJob)
}

func (cjt *CrawlJobsTest) testAddCrawlJobReturnsBadRequestOnInvalidJson(t *testing.T) {

	reader := strings.NewReader(`{'test':'test'}`)
	resp, err := http.Post(
		cjt.server.URL+"/crawlJobs",
		"application/json",
		reader)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func (cjt *CrawlJobsTest) testAddCrawlJobReturnsJobIdIfAlreadyAdded(t *testing.T) {

	job := dal.CrawlJob{
		BaseUrl:     "test",
		LastUpdated: "10:11:10",
		Status:      dal.InProgress,
		JobId:       123,
	}

	request := CrawlJobRequest{
		BaseUrl: "test",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)

	if err != nil {
		t.Fatal(err)
	}

	cjt.mockCrawlJobRepo.EXPECT().GetCrawlJobForUrl("test").Return(job, nil)
	resp, err := http.Post(
		cjt.server.URL+"/crawlJobs",
		"application/json",
		&buf)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//read the body
	buf.Reset()
	buf.ReadFrom(resp.Body)
	result := buf.String()

	assert.Equal(t, "123\n", result)
}

func (cjt *CrawlJobsTest) testSuccessfulAddCrawlJob(t *testing.T) {

	var discoveredLinks map[string]struct{}
	cjt.mockCrawlJobRepo.EXPECT().GetCrawlJobForUrl("test").Return(dal.CrawlJob{}, nil)
	cjt.mockCrawlJobRepo.EXPECT().AddCrawlJob("test").Return(123, nil)
	cjt.mockCrawlJobRepo.EXPECT().UpdateCrawlJobStatus(123, dal.Completed).Return(nil)
	cjt.mockWebCrawler.EXPECT().Crawl("test", gomock.Any()).Return(discoveredLinks, nil).Times(1)

	request := CrawlJobRequest{
		BaseUrl: "test",
	}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)

	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.Post(
		cjt.server.URL+"/crawlJobs",
		"application/json",
		&buf)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	//read the body
	buf.Reset()
	buf.ReadFrom(resp.Body)
	result := buf.String()

	assert.Equal(t, "123\n", result)
}
