package crawler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type LinkCrawlerTest struct {
	crawler *LinkCrawler
	server  *httptest.Server
	mux     *http.ServeMux
}

func TestLinkCrawler(t *testing.T) {

	lct := &LinkCrawlerTest{}
	teardown := lct.setupSuite(t)
	defer teardown(t)

	t.Run("Test returns empty hashset on status not ok", lct.testReturnsEmptyHashsetOnStatusNotOK)
	t.Run("Test returns empty hashset if no links", lct.testReturnsEmptyHashsetIfNoLinks)
	t.Run("Test successful crawl", lct.testSuccessfulCrawl)
}

func (lct *LinkCrawlerTest) setupSuite(t *testing.T) func(t *testing.T) {
	client := &http.Client{}
	lct.crawler = NewCrawler(
		client,
		NewPolicyExecutor("//a[@href[not(contains(.,'http')) and not(contains(.,'mailto:')) and not(contains(.,'tel:'))]]"))

	return func(t *testing.T) {
		client.CloseIdleConnections()
	}
}

func (lct *LinkCrawlerTest) setupTest(t *testing.T) func(t *testing.T) {
	mux := http.NewServeMux()
	lct.mux = mux
	lct.server = httptest.NewServer(mux)
	return func(t *testing.T) {
		lct.server.Close()
	}
}

func (lct *LinkCrawlerTest) setupMockHandler(t *testing.T, contentMap map[string]string) {

	for key := range contentMap {
		pattern := fmt.Sprintf("/%s", key)
		currentContent := contentMap[key]
		lct.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(currentContent))
		})
	}
}

func (lct *LinkCrawlerTest) setupMockHandlerWithError(content string, code int) {
	lct.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, content, code)
	})
}

func (lct *LinkCrawlerTest) testReturnsEmptyHashsetOnStatusNotOK(t *testing.T) {
	td := lct.setupTest(t)
	defer td(t)
	lct.setupMockHandlerWithError("some error", 401)
	baseUrl := lct.server.URL
	hashset, err := lct.crawler.Crawl(baseUrl, func(links []string) error { return nil })

	assert.Nil(t, err)
	assert.Len(t, hashset, 0)

}

func (lct *LinkCrawlerTest) testReturnsEmptyHashsetIfNoLinks(t *testing.T) {
	td := lct.setupTest(t)
	defer td(t)

	htmlContent1 := `
		<html>
			<div>test</div>
		</html>`

	contentMap := map[string]string{
		"test": htmlContent1,
	}

	lct.setupMockHandler(t, contentMap)

	baseUrl := lct.server.URL
	hashset, err := lct.crawler.Crawl(baseUrl+"/test", func(links []string) error { return nil })

	assert.Nil(t, err)
	assert.Len(t, hashset, 0)
}

func (lct *LinkCrawlerTest) testSuccessfulCrawl(t *testing.T) {
	td := lct.setupTest(t)
	defer td(t)

	htmlContent0 := `
		<html>
			<a href='/test1'>test1</a>
		</html>`
	htmlContent1 := `
		<html>
			<a href='/test2'>test2</a>
		</html>`
	htmlContent2 := `
		<html>
			<a href='/test1'>test1</a>
			<a href='/test3'>test3</a>
		</html>`
	htmlContent3 := `<html>
		<a href='/test4'>test4</a>
	</html>`
	htmlContent4 := `<html>
		<a href='/test2'>test2</a>
		<a href='/test3'>test3</a>
	</html>`

	contentMap := map[string]string{
		"":      htmlContent0,
		"test1": htmlContent1,
		"test2": htmlContent2,
		"test3": htmlContent3,
		"test4": htmlContent4,
	}

	lct.setupMockHandler(t, contentMap)

	baseUrl := lct.server.URL
	discoveredLinks, err := lct.crawler.Crawl(baseUrl, func(links []string) error { return nil })

	assert.Nil(t, err)
	assert.Len(t, discoveredLinks, 4)
}
