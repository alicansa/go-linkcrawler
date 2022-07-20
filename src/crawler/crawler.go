package crawler

//go:generate mockgen -destination=../mocks/mock_crawler.go -package=mocks github.com/alicansa/go-linkcrawler/crawler CrawlPolicyExecuter,WebCrawler

import "io"

type CrawlPolicyExecuter interface {
	Execute(rc io.ReadCloser) ([]string, error)
}

type WebCrawler interface {
	Crawl(url string, onLinksDiscovered func(links []string) error) (map[string]struct{}, error)
}
