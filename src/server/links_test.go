package server

import (
	"encoding/json"
	"errors"
	"testing"

	"net/http"
	"net/http/httptest"

	"github.com/alicansa/go-linkcrawler/dal"
	"github.com/alicansa/go-linkcrawler/mocks"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

type LinksTest struct {
	linksRepo  *mocks.MockLinkRepository
	server     *httptest.Server
	controller *gomock.Controller
}

func TestLinks(t *testing.T) {
	// setup + teardown
	lt := &LinksTest{}
	tds := lt.setupSuite(t)
	defer tds(t)

	t.Run("Test returns bad request on invalid job id", lt.testInvalidJobId)
	t.Run("Test returns internal server error on db error", lt.testLinkRepositoryError)
	t.Run("Test successfully returns links", lt.testSuccessfulGetLinks)
}

func (lt *LinksTest) setupSuite(t *testing.T) func(t *testing.T) {
	lt.controller = gomock.NewController(t)
	mockLinkRepo := mocks.NewMockLinkRepository(lt.controller)
	lt.linksRepo = mockLinkRepo
	r := mux.NewRouter()
	linksHandler := NewLinksHandler(mockLinkRepo)
	linksHandler.registerLinksHandler(r)
	lt.server = httptest.NewServer(r)

	return func(t *testing.T) {
		lt.server.Close()
	}
}

func (lt *LinksTest) testInvalidJobId(t *testing.T) {

	res, err := http.Get(lt.server.URL + "/links?crawlJobId=test")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func (lt *LinksTest) testLinkRepositoryError(t *testing.T) {

	lt.linksRepo.EXPECT().GetLinks(123).Return(nil, errors.New("db error")).Times(1)

	res, err := http.Get(lt.server.URL + "/links?crawlJobId=123")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func (lt *LinksTest) testSuccessfulGetLinks(t *testing.T) {

	expectedLinks := []dal.Link{
		{Url: "test.com/test", LinkId: 12345, CrawlJobId: 123},
	}
	lt.linksRepo.EXPECT().GetLinks(123).Return(expectedLinks, nil).Times(1)

	res, err := http.Get(lt.server.URL + "/links?crawlJobId=123")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)

	//check header content type
	assert.Equal(t, "application/json", res.Header.Get("Content-type"))

	//check response
	var decodedLinks []dal.Link
	if err := json.NewDecoder(res.Body).Decode(&decodedLinks); err != nil {
		t.Fatal(err)
	}

	assert.Len(t, decodedLinks, 1)
	assert.Equal(t, expectedLinks, decodedLinks)
}
