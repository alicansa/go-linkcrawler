package dal

//go:generate mockgen -destination=../mocks/mock_dal.go -package=mocks github.com/alicansa/go-linkcrawler/dal LinkRepository,CrawlJobRepository
//go:generate stringer -type=CrawlJobStatus

type Link struct {
	Url        string `json:"url"`
	LinkId     int    `json:"linkId"`
	CrawlJobId int    `json:"crawlJobId"`
}

type CrawlJobStatus int

const (
	InProgress CrawlJobStatus = iota + 1
	Completed
)

type CrawlJob struct {
	LastUpdated string         `json:"lastUpdated"`
	BaseUrl     string         `json:"baseUrl"`
	Status      CrawlJobStatus `json:"status"`
	JobId       int            `json:"jobId"`
}

type LinkRepository interface {
	GetLinks(crawlJobId int) ([]Link, error)
	AddLink(url string, crawlJobId int) (int, error)
}

type CrawlJobRepository interface {
	AddCrawlJob(baseUrl string) (int, error)
	UpdateCrawlJobStatus(crawlJobId int, status CrawlJobStatus) error
	GetCrawlJob(crawlJobId int) (CrawlJob, error)
	GetCrawlJobForUrl(url string) (CrawlJob, error)
	GetCrawlJobs() ([]CrawlJob, error)
}
