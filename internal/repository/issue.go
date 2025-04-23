package repository

import (
	"database/sql"
	"log"
	"math"

	"github.com/stjudewashere/seonaut/internal/models"
)

type IssueRepository struct {
	DB *sql.DB
}

// SaveIssues inserts the issues it receives in the iStream channel into the database
// using a batch process.
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

// FindIssuesByTypeAndPriority returns an IssueGroup model with all the issues detected in a crawl
// with the specified priority and categorized by error type.
func (ds *IssueRepository) FindIssuesByTypeAndPriority(cid int64, p int) []models.IssueGroup {
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

// FindPassedIssues returns an IssueGroup model with all the issues types that have passed
// and don't have any reported issue for the specified crawl.
func (ds *IssueRepository) FindPassedIssues(cid int64) []models.IssueGroup {
	issues := []models.IssueGroup{}
	query := `
		SELECT
			issue_types.type,
			issue_types.priority,
			count(DISTINCT issues.pagereport_id) AS c
		FROM issue_types
		LEFT JOIN  issues ON issue_types.id = issues.issue_type_id AND issues.crawl_id = ?
		GROUP BY issue_types.id, issue_types.type, issue_types.priority
		HAVING COUNT(issues.id) = 0
		ORDER BY issue_types.type;`

	rows, err := ds.DB.Query(query, cid)
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

// CountIssuesByPriority returns the total number of issues of the specified priority
// found in a crawl.
func (ds *IssueRepository) CountIssuesByPriority(cid int64, p int) int {
	query := `
		SELECT
			count(issues.pagereport_id) AS c
		FROM issues
		INNER JOIN  issue_types ON issue_types.id = issues.issue_type_id
		WHERE crawl_id = ? AND issue_types.priority = ? GROUP BY issue_types.priority`

	row := ds.DB.QueryRow(query, cid, p)
	var c int
	if err := row.Scan(&c); err != nil && err != sql.ErrNoRows {
		log.Printf("CountIssuesByPriority: %v\n", err)
	}

	return c
}

// GetNumberOfPagesForIssues returns the total number of pages for an specific issue "errorType". This can
// be used in combination with FindPageReportIssues to generate a paginated view of the issues.
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

// FindPageReportIssues returns a slice of PageReports corresponding to the page specified in the "p" parameter
// and with the errorType specified in "errorType".
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

// Return the issue types found for an specific page report.
func (ds *IssueRepository) FindErrorTypesByPage(pid int, cid int64) []string {
	var et []string
	query := `
		SELECT 
			issue_types.type
		FROM issues
		INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
		WHERE pagereport_id = ? and crawl_id = ?
		GROUP BY issue_type_id`

	rows, err := ds.DB.Query(query, pid, cid)
	if err != nil {
		log.Println(err)
		return et
	}

	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			log.Println(err)
			continue
		}
		et = append(et, s)
	}

	return et
}
