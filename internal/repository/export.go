package repository

import (
	"database/sql"
	"log"

	"github.com/stjudewashere/seonaut/internal/models"
)

type ExportRepository struct {
	DB *sql.DB
}

// Send all internal links through a read-only channel
func (ds *ExportRepository) ExportLinks(crawl *models.Crawl) <-chan *models.ExportLink {
	lStream := make(chan *models.ExportLink)

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

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.ExportLink{}
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
func (ds *ExportRepository) ExportExternalLinks(crawl *models.Crawl) <-chan *models.ExportLink {
	lStream := make(chan *models.ExportLink)

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

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.ExportLink{}
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
func (ds *ExportRepository) ExportImages(crawl *models.Crawl) <-chan *models.ExportImage {
	iStream := make(chan *models.ExportImage)

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

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.ExportImage{}
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

// Send all scripts URLs through a read-only channel
func (ds *ExportRepository) ExportScripts(crawl *models.Crawl) <-chan *models.Script {
	sStream := make(chan *models.Script)

	go func() {
		defer close(sStream)

		query := `
			SELECT
				pagereports.url,
				scripts.url
			FROM scripts
			LEFT JOIN pagereports ON pagereports.id  = scripts.pagereport_id
			WHERE scripts.crawl_id = ?`

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.Script{}
			err := rows.Scan(&v.Origin, &v.Script)
			if err != nil {
				log.Println(err)
				continue
			}

			sStream <- v
		}
	}()

	return sStream
}

// Send all css style URLs through a read-only channel
func (ds *ExportRepository) ExportStyles(crawl *models.Crawl) <-chan *models.Style {
	sStream := make(chan *models.Style)

	go func() {
		defer close(sStream)

		query := `
			SELECT
				pagereports.url,
				styles.url
			FROM styles
			LEFT JOIN pagereports ON pagereports.id  = styles.pagereport_id
			WHERE styles.crawl_id = ?`

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.Style{}
			err := rows.Scan(&v.Origin, &v.Style)
			if err != nil {
				log.Println(err)
				continue
			}

			sStream <- v
		}
	}()

	return sStream
}

// Send all iframe URLs through a read-only channel
func (ds *ExportRepository) ExportIframes(crawl *models.Crawl) <-chan *models.Iframe {
	vStream := make(chan *models.Iframe)

	go func() {
		defer close(vStream)

		query := `
			SELECT
				pagereports.url,
				iframes.url
			FROM iframes
			LEFT JOIN pagereports ON pagereports.id  = iframes.pagereport_id
			WHERE iframes.crawl_id = ?`

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.Iframe{}
			err := rows.Scan(&v.Origin, &v.Iframe)
			if err != nil {
				log.Println(err)
				continue
			}

			vStream <- v
		}
	}()

	return vStream
}

// Send all audio URLs through a read-only channel
func (ds *ExportRepository) ExportAudios(crawl *models.Crawl) <-chan *models.Audio {
	vStream := make(chan *models.Audio)

	go func() {
		defer close(vStream)

		query := `
			SELECT
				pagereports.url,
				audios.url
			FROM audios
			LEFT JOIN pagereports ON pagereports.id  = audios.pagereport_id
			WHERE audios.crawl_id = ?`

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.Audio{}
			err := rows.Scan(&v.Origin, &v.Audio)
			if err != nil {
				log.Println(err)
				continue
			}

			vStream <- v
		}
	}()

	return vStream
}

// Send all video URLs through a read-only channel
func (ds *ExportRepository) ExportVideos(crawl *models.Crawl) <-chan *models.ExportVideo {
	vStream := make(chan *models.ExportVideo)

	go func() {
		defer close(vStream)

		query := `
			SELECT
				pagereports.url,
				videos.url
			FROM videos
			LEFT JOIN pagereports ON pagereports.id  = videos.pagereport_id
			WHERE videos.crawl_id = ?`

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.ExportVideo{}
			err := rows.Scan(&v.Origin, &v.Video)
			if err != nil {
				log.Println(err)
				continue
			}

			vStream <- v
		}
	}()

	return vStream
}

// Send all hreflang URLs and language through a read-only channel
func (ds *ExportRepository) ExportHreflangs(crawl *models.Crawl) <-chan *models.ExportHreflang {
	vStream := make(chan *models.ExportHreflang)

	go func() {
		defer close(vStream)

		query := `
			SELECT
				pagereports.url,
				hreflangs.from_lang,
				hreflangs.to_url,
				hreflangs.to_lang
			FROM hreflangs
			LEFT JOIN pagereports ON pagereports.id  = hreflangs.pagereport_id
			WHERE hreflangs.crawl_id = ?`

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.ExportHreflang{}
			err := rows.Scan(&v.Origin, &v.OriginLang, &v.Hreflang, &v.HreflangLang)
			if err != nil {
				log.Println(err)
				continue
			}

			vStream <- v
		}
	}()

	return vStream
}

// Export all issues by crawl through a read-only channel
func (ds *ExportRepository) ExportIssues(crawl *models.Crawl) <-chan *models.ExportIssue {
	vStream := make(chan *models.ExportIssue)

	go func() {
		defer close(vStream)

		query := `
		SELECT
			pagereports.url,
			issue_types.type,
			issue_types.priority
		FROM issues
			LEFT JOIN  issue_types ON issue_types.id = issues.issue_type_id
			LEFT JOIN pagereports ON pagereports.id = issues.pagereport_id
		WHERE issues.crawl_id = ?
		ORDER BY issue_types.priority ASC`

		rows, err := ds.DB.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &models.ExportIssue{}
			err := rows.Scan(&v.Url, &v.Type, &v.Priority)
			if err != nil {
				log.Println(err)
				continue
			}

			vStream <- v
		}
	}()

	return vStream
}
