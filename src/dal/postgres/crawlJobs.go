package postgres

import (
	"database/sql"
	"time"

	"github.com/alicansa/go-linkcrawler/dal"
)

type CrawlJobRepository struct {
	db *DB
}

func (cjr *CrawlJobRepository) AddCrawlJob(baseUrl string) (int, error) {
	var linkId int
	sqlStatement := `
		INSERT INTO crawljob (crawljobstatus_id, base_url, last_updated)
		VALUES ($1, $2, $3)
		RETURNING job_id`

	err := cjr.db.db.QueryRow(
		sqlStatement,
		dal.InProgress,
		baseUrl,
		time.Now().UTC()).Scan(&linkId)

	if err != nil {
		return linkId, err
	}

	return linkId, nil
}

func (repo *CrawlJobRepository) UpdateCrawlJobStatus(jobId int, status dal.CrawlJobStatus) error {
	sqlStatement := `
		UPDATE crawljob 
		SET crawljobstatus_id = $1
		WHERE job_id = $2`

	_, err := repo.db.db.Exec(sqlStatement, status, jobId)

	return err
}

func (repo *CrawlJobRepository) GetCrawlJob(crawlJobId int) (dal.CrawlJob, error) {
	var jobId int
	var baseUrl string
	var jobStatus dal.CrawlJobStatus
	var lastUpdated string

	err := repo.db.db.QueryRow(
		`SELECT job_id, crawljobstatus_id, base_url, last_updated FROM crawljob WHERE job_id=$1`,
		crawlJobId).Scan(&jobId, &jobStatus, &baseUrl, &lastUpdated)

	if err != nil {

		if err == sql.ErrNoRows {
			return dal.CrawlJob{}, nil
		}

		return dal.CrawlJob{}, err

	}

	return dal.CrawlJob{
		LastUpdated: lastUpdated,
		BaseUrl:     baseUrl,
		Status:      jobStatus,
		JobId:       jobId,
	}, nil
}

func (repo *CrawlJobRepository) GetCrawlJobForUrl(url string) (dal.CrawlJob, error) {
	var jobId int
	var baseUrl string
	var jobStatus dal.CrawlJobStatus
	var lastUpdated string

	err := repo.db.db.QueryRow(
		`SELECT job_id, crawljobstatus_id, base_url, last_updated FROM crawljob WHERE base_url=$1`,
		url).Scan(&jobId, &jobStatus, &baseUrl, &lastUpdated)

	if err != nil {

		if err == sql.ErrNoRows {
			return dal.CrawlJob{}, nil
		}

		return dal.CrawlJob{}, err

	}

	return dal.CrawlJob{
		LastUpdated: lastUpdated,
		BaseUrl:     baseUrl,
		Status:      jobStatus,
		JobId:       jobId,
	}, nil
}

func (repo *CrawlJobRepository) GetCrawlJobs() ([]dal.CrawlJob, error) {
	var jobs []dal.CrawlJob
	rows, err := repo.db.db.Query(`SELECT job_id, crawljobstatus_id, base_url, last_updated FROM crawljob`)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var jobId int
		var baseUrl string
		var jobStatus dal.CrawlJobStatus
		var lastUpdated string

		err = rows.Scan(&jobId, &jobStatus, &baseUrl, &lastUpdated)

		if err != nil {
			return nil, err
		}

		jobs = append(jobs, dal.CrawlJob{
			LastUpdated: lastUpdated,
			BaseUrl:     baseUrl,
			Status:      jobStatus,
			JobId:       jobId,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func NewCrawlJobRepository(db *DB) *CrawlJobRepository {
	return &CrawlJobRepository{db: db}
}
