package sql_reporters

import (
	"database/sql"
	"log"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/report_manager"
)

type SqlReporter struct {
	db *sql.DB
}

// NewSqlReporter creates a new SqlReporter with the given SQL database connection.
func NewSqlReporter(db *sql.DB) *SqlReporter {
	return &SqlReporter{
		db: db,
	}
}

// GetAllReporters returns a slice of all the implemented reporters in the SqlReporter.
func (sr *SqlReporter) GetAllReporters() []report_manager.MultipageCallback {
	return []report_manager.MultipageCallback{
		// Add status code issue reporters
		sr.RedirectChainsReporter,
		sr.RedirectLoopsReporter,

		// Add title issue reporters
		sr.DuplicatedTitleReporter,

		// Add description issue reporters
		sr.DuplicatedDescriptionReporter,

		// Add link issue reporters
		sr.OrphanPagesReporter,
		sr.NoFollowIndexableReporter,
		sr.FollowNoFollowReporter,

		// Add hreflang reporters
		sr.MissingHrelangReturnLinks,
		sr.HreflangsToNonCanonical,
		sr.HreflangNoindexable,

		// Add canonical issue reporters
		sr.CanonicalizedToNonCanonical,
	}
}

// pageReportsQuery executes a SQL query and returns a channel of *models.PageReport.
func (sr *SqlReporter) pageReportsQuery(query string, args ...interface{}) <-chan *models.PageReport {
	prStream := make(chan *models.PageReport)

	go func() {
		defer close(prStream)

		rows, err := sr.db.Query(query, args...)
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
