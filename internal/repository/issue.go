package repository

import (
	"database/sql"
	"log"
	"math"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
)

type IssueRepository struct {
	DB *sql.DB
}

func (ds *IssueRepository) SaveEndIssues(cid int64, t time.Time) {
	stmt, _ := ds.DB.Prepare("UPDATE crawls SET issues_end = ? WHERE id = ?")
	defer stmt.Close()
	_, err := stmt.Exec(t, cid)
	if err != nil {
		log.Printf("saveEndIssues: %v\n", err)
	}
}

func (ds *IssueRepository) SaveIssues(iStream <-chan *models.Issue) {
	query := "INSERT INTO issues (pagereport_id, crawl_id, issue_type_id) VALUES "
	sqlString := ""
	v := []interface{}{}

	fn := func() {
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := ds.DB.Prepare(query + sqlString)
		defer stmt.Close()

		_, err := stmt.Exec(v...)
		if err != nil {
			log.Println(err)
		}

		v = []interface{}{}
		sqlString = ""
	}

	for i := range iStream {
		sqlString += "(?, ?, ?),"
		v = append(v, i.PageReportId, i.CrawlId, i.ErrorType)

		if len(v) >= 100 {
			fn()
		}
	}

	if len(v) > 0 {
		fn()
	}
}

func (ds *IssueRepository) FindIssuesByPriority(cid int64, p int) []models.IssueGroup {
	issues := []models.IssueGroup{}
	query := `
		SELECT
			issue_types.type,
			issue_types.priority,
			count(DISTINCT issues.pagereport_id) AS c
		FROM issues
		INNER JOIN  issue_types ON issue_types.id = issues.issue_type_id
		WHERE crawl_id = ? AND issue_types.priority = ? GROUP BY issue_type_id
		ORDER BY c DESC`

	rows, err := ds.DB.Query(query, cid, p)
	if err != nil {
		log.Println(err)
		return issues
	}

	for rows.Next() {
		ig := models.IssueGroup{}
		err := rows.Scan(&ig.ErrorType, &ig.Priority, &ig.Count)
		if err != nil {
			log.Println(err)
			continue
		}

		issues = append(issues, ig)
	}

	return issues
}

func (ds *IssueRepository) GetNumberOfPagesForIssues(cid int64, errorType string) int {
	query := `
		SELECT count(DISTINCT pagereport_id)
		FROM issues
		INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
		WHERE issue_types.type = ? AND crawl_id  = ?`

	row := ds.DB.QueryRow(query, errorType, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForIssues: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}

func (ds *IssueRepository) FindPageReportIssues(cid int64, p int, errorType string) []models.PageReport {
	max := paginationMax
	offset := max * (p - 1)

	query := `
		SELECT
			id,
			url,
			title
		FROM pagereports
		WHERE id IN (
			SELECT DISTINCT pagereport_id
			FROM issues
			INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
			WHERE issue_types.type = ? AND crawl_id = ?
		) ORDER BY url ASC LIMIT ?, ?`

	var pageReports []models.PageReport
	rows, err := ds.DB.Query(query, errorType, cid, offset, max)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		p := models.PageReport{}
		err := rows.Scan(&p.Id, &p.URL, &p.Title)
		if err != nil {
			log.Println(err)
			continue
		}

		pageReports = append(pageReports, p)
	}

	return pageReports
}

func (ds *IssueRepository) SaveIssuesCount(crawlId int64, critical, alert, warning int) {
	query := `UPDATE
		crawls
		SET 
			critical_issues = ?,
			alert_issues = ?,
			warning_issues = ?,
			total_issues = ?
		WHERE id = ?`

	stmt, _ := ds.DB.Prepare(query)
	defer stmt.Close()

	total := critical + alert + warning
	_, err := stmt.Exec(critical, alert, warning, total, crawlId)
	if err != nil {
		log.Printf("SaveIssuesCount: %v\n", err)
	}
}
