package postgres

import (
	"github.com/alicansa/go-linkcrawler/dal"
)

type LinkRepository struct {
	db *DB
}

func (lr *LinkRepository) GetLinks(crawlJobId int) ([]dal.Link, error) {

	var links []dal.Link
	rows, err := lr.db.db.Query(`SELECT link_id, url FROM crawllink WHERE crawljob_id=$1`, crawlJobId)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	// loops over rows and add to links

	for rows.Next() {
		var linkId int
		var url string

		err = rows.Scan(&linkId, &url)

		if err != nil {
			// handle this error
			return nil, err
		}

		links = append(links, dal.Link{
			Url:        url,
			CrawlJobId: crawlJobId,
			LinkId:     linkId,
		})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (lr *LinkRepository) AddLink(url string, crawlJobId int) (int, error) {

	var linkId int
	sqlStatement := `
		INSERT INTO crawllink (url, crawljob_id)
		VALUES ($1, $2)
		RETURNING link_id`

	err := lr.db.db.QueryRow(sqlStatement, url, crawlJobId).Scan(&linkId)

	if err != nil {
		return linkId, err
	}

	return linkId, nil
}

func NewLinkRepository(db *DB) *LinkRepository {
	return &LinkRepository{db: db}
}
