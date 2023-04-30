package datastore

import (
	"log"

	"github.com/stjudewashere/seonaut/internal/models"
)

func (ds *Datastore) PageReportsQuery(query string, args ...interface{}) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)

	go func() {
		defer close(prStream)

		rows, err := ds.db.Query(query, args...)
		if err != nil {
			log.Println(err)
		}

		for rows.Next() {
			p := &models.PageReport{}
			err := rows.Scan(&p.Id, &p.URL, &p.Title)
			if err != nil {
				log.Println(err)
				continue
			}

			prStream <- p
		}
	}()

	return prStream
}
