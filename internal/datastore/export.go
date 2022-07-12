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

// Send all scripts URLs through a read-only channel
func (ds *Datastore) ExportScripts(crawl *crawler.Crawl) <-chan *export.Script {
	sStream := make(chan *export.Script)

	go func() {
		defer close(sStream)

		query := `
			SELECT
				pagereports.url,
				scripts.url
			FROM scripts
			LEFT JOIN pagereports ON pagereports.id  = scripts.pagereport_id
			WHERE scripts.crawl_id = ?`

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Script{}
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
func (ds *Datastore) ExportStyles(crawl *crawler.Crawl) <-chan *export.Style {
	sStream := make(chan *export.Style)

	go func() {
		defer close(sStream)

		query := `
			SELECT
				pagereports.url,
				styles.url
			FROM styles
			LEFT JOIN pagereports ON pagereports.id  = styles.pagereport_id
			WHERE styles.crawl_id = ?`

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Style{}
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
func (ds *Datastore) ExportIframes(crawl *crawler.Crawl) <-chan *export.Iframe {
	vStream := make(chan *export.Iframe)

	go func() {
		defer close(vStream)

		query := `
			SELECT
				pagereports.url,
				iframes.url
			FROM iframes
			LEFT JOIN pagereports ON pagereports.id  = iframes.pagereport_id
			WHERE iframes.crawl_id = ?`

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Iframe{}
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
func (ds *Datastore) ExportAudios(crawl *crawler.Crawl) <-chan *export.Audio {
	vStream := make(chan *export.Audio)

	go func() {
		defer close(vStream)

		query := `
			SELECT
				pagereports.url,
				audios.url
			FROM audios
			LEFT JOIN pagereports ON pagereports.id  = audios.pagereport_id
			WHERE audios.crawl_id = ?`

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Audio{}
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
func (ds *Datastore) ExportVideos(crawl *crawler.Crawl) <-chan *export.Video {
	vStream := make(chan *export.Video)

	go func() {
		defer close(vStream)

		query := `
			SELECT
				pagereports.url,
				videos.url
			FROM videos
			LEFT JOIN pagereports ON pagereports.id  = videos.pagereport_id
			WHERE videos.crawl_id = ?`

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Video{}
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
func (ds *Datastore) ExportHreflangs(crawl *crawler.Crawl) <-chan *export.Hreflang {
	vStream := make(chan *export.Hreflang)

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

		rows, err := ds.db.Query(query, crawl.Id)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			v := &export.Hreflang{}
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
