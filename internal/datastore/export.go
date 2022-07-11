package datastore

import (
	"log"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/export"
)

// Send all internal links through a read-only channel
func (ds *Datastore) ExportLinks(crawl *crawler.Crawl) <-chan *export.Link {
	lStream := make(chan *export.Link)

	go func() {
		defer close(lStream)

		query := `
				SELECT
					pagereports.url,
					links.url,
					links.text
				FROM links
				LEFT JOIN pagereports ON pagereports.id  = links.pagereport_id
				WHERE links.crawl_id = ?`

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Link{}
			err := rows.Scan(&v.Origin, &v.Destination, &v.Text)
			if err != nil {
				log.Println(err)
				continue
			}

			lStream <- v
		}
	}()

	return lStream
}

// Send all external links through a read-only channel
func (ds *Datastore) ExportExternalLinks(crawl *crawler.Crawl) <-chan *export.Link {
	lStream := make(chan *export.Link)

	go func() {
		defer close(lStream)

		query := `
				SELECT
					pagereports.url,
					external_links.url,
					external_links.text
				FROM external_links
				LEFT JOIN pagereports ON pagereports.id  = external_links.pagereport_id
				WHERE external_links.crawl_id = ?`

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Link{}
			err := rows.Scan(&v.Origin, &v.Destination, &v.Text)
			if err != nil {
				log.Println(err)
				continue
			}

			lStream <- v
		}
	}()

	return lStream
}

// Send all image URLs through a read-only channel
func (ds *Datastore) ExportImages(crawl *crawler.Crawl) <-chan *export.Image {
	iStream := make(chan *export.Image)

	go func() {
		defer close(iStream)

		query := `
			SELECT
				pagereports.url,
				images.url,
				images.alt
			FROM images
			LEFT JOIN pagereports ON pagereports.id  = images.pagereport_id
			WHERE images.crawl_id = ?`

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Image{}
			err := rows.Scan(&v.Origin, &v.Image, &v.Alt)
			if err != nil {
				log.Println(err)
				continue
			}

			iStream <- v
		}
	}()

	return iStream
}
