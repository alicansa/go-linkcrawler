package crawler

import (
	"context"
	"net/http"
	"strings"
	"sync"
)

const (
	MaxNumberOfCrawlers = 10
)

type LinkCrawler struct {
	Client          *http.Client
	discoveredLinks threadSafeHashSet
	PolicyExecuter  CrawlPolicyExecuter
}

func (c *LinkCrawler) Crawl(url string, onLinksDiscovered func(links []string) error) (map[string]struct{}, error) {

	// create a new thread safe hashset for this crawl
	c.discoveredLinks = threadSafeHashSet{
		hashset: make(map[string]struct{}),
		mx:      &sync.Mutex{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	links := []string{url}
	err := c.crawlRecursive(
		url,
		"",
		links,
		ctx,
		cancel,
		onLinksDiscovered)

	if err != nil {
		return c.discoveredLinks.hashset, err
	}

	return c.discoveredLinks.hashset, nil
}

func (c *LinkCrawler) crawlRecursive(
	baseUrl string,
	relativeUrl string,
	links []string,
	ctx context.Context,
	cancel context.CancelFunc,
	onLinksDiscovered func(links []string) error) error {
	// if there are no links to be crawled, just return
	if len(links) == 0 {
		return nil
	}

	workerChan := make(chan int, MaxNumberOfCrawlers)
	resultChan := make(chan getLinksResult, len(links))
	errChan := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(len(links))

	for _, l := range links {
		// if link starts with # just continue
		if len(l) == 0 || l[0] == '#' {
			resultChan <- getLinksResult{}
			wg.Done()
			continue
		}

		//block if max number of crawlers already crawling
		workerChan <- 1
		go func(link string) {
			defer wg.Done()

			select {
			//in case another worker cancels context
			case <-ctx.Done():
				resultChan <- getLinksResult{}
				<-workerChan
				return
			default:
			}

			processedLink := processLink(link, relativeUrl, baseUrl)
			resp, err := c.getLinks(baseUrl + processedLink)
			resp.relativeUrl = processedLink

			if err != nil {
				errChan <- err
				cancel()
				close(workerChan)
				close(resultChan)
				return
			}
			//if there are no links crawler hasn't visited, just return
			if len(resp.links) == 0 {
				resultChan <- resp
				// release channel
				<-workerChan
				return
			}

			//callback function for discovered links
			var newLinks []string
			for _, discoveredLink := range resp.links {

				if !c.discoveredLinks.Add(discoveredLink) {
					continue
				}

				newLinks = append(newLinks, discoveredLink)
			}

			err = onLinksDiscovered(newLinks)

			if err != nil {
				errChan <- err
				cancel()
				close(workerChan)
				close(resultChan)
				return
			}
			// release channel
			<-workerChan
			// return on results channel
			resultChan <- getLinksResult{
				links:       newLinks,
				relativeUrl: processedLink,
			}
		}(l)
	}

	wg.Wait()
	close(resultChan)

	select {
	case err := <-errChan:
		return err
	default:
		for result := range resultChan {
			err := c.crawlRecursive(baseUrl, result.relativeUrl, result.links, ctx, cancel, onLinksDiscovered)

			if err != nil {
				return err
			}
		}
	}
	return nil
}

type getLinksResult struct {
	relativeUrl string
	links       []string
}

func (c *LinkCrawler) getLinks(url string) (getLinksResult, error) {

	resp, err := c.Client.Get(url)

	if err != nil || resp.StatusCode != http.StatusOK {
		return getLinksResult{}, err
	}

	defer resp.Body.Close()

	links, err := c.PolicyExecuter.Execute(resp.Body)

	if err != nil {
		return getLinksResult{}, err
	}

	return getLinksResult{
		links: links,
	}, nil
}

func processLink(
	link string,
	relativeUrl string,
	baseUrl string) string {

	if link[0] == '#' {
		return relativeUrl + link
	}

	if link == baseUrl {
		return ""
	}

	if strings.Contains(link, baseUrl) {
		return link
	}

	httpEndIndex := strings.Index(baseUrl, "//")
	if httpEndIndex >= 0 && strings.Contains(link, baseUrl[0:httpEndIndex+2]) {
		var startIndex = strings.Index(link, "/")
		return link[0:startIndex]
	}

	if link[0] == '/' {
		return link
	}

	return relativeUrl + "/" + link
}

func NewCrawler(
	httpClient *http.Client,
	pe CrawlPolicyExecuter) *LinkCrawler {
	return &LinkCrawler{
		Client:         httpClient,
		PolicyExecuter: pe,
	}
}

type threadSafeHashSet struct {
	mx      *sync.Mutex
	hashset map[string]struct{}
}

func (tsh *threadSafeHashSet) Exists(key string) bool {
	_, ok := tsh.hashset[key]
	return ok
}

func (tsh *threadSafeHashSet) Add(key string) bool {
	//check exists
	if tsh.Exists(key) {
		return false
	}
	//lock
	tsh.mx.Lock()
	//add
	tsh.hashset[key] = struct{}{}
	//unlock
	tsh.mx.Unlock()
	return true
}
