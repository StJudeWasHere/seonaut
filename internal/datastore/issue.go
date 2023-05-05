package datastore

import (
	"log"
	"math"
	"time"

	"github.com/stjudewashere/seonaut/internal/issue"
	"github.com/stjudewashere/seonaut/internal/models"
)

func (ds *Datastore) SaveEndIssues(cid int64, t time.Time) {
	stmt, _ := ds.db.Prepare("UPDATE crawls SET issues_end = ? WHERE id = ?")
	defer stmt.Close()
	_, err := stmt.Exec(t, cid)
	if err != nil {
		log.Printf("saveEndIssues: %v\n", err)
	}
}

func (ds *Datastore) SaveIssues(iStream <-chan *models.Issue) {
	query := "INSERT INTO issues (pagereport_id, crawl_id, issue_type_id) VALUES "
	sqlString := ""
	v := []interface{}{}

	fn := func() {
		sqlString = sqlString[0 : len(sqlString)-1]
		stmt, _ := ds.db.Prepare(query + sqlString)
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

func (ds *Datastore) FindIssuesByPriority(cid int64, p int) []issue.IssueGroup {
	issues := []issue.IssueGroup{}
	query := `
		SELECT
			issue_types.type,
			issue_types.priority,
			count(DISTINCT issues.pagereport_id) AS c
		FROM issues
		INNER JOIN  issue_types ON issue_types.id = issues.issue_type_id
		WHERE crawl_id = ? AND issue_types.priority = ? GROUP BY issue_type_id
		ORDER BY c DESC`

	rows, err := ds.db.Query(query, cid, p)
	if err != nil {
		log.Println(err)
		return issues
	}

	for rows.Next() {
		ig := issue.IssueGroup{}
		err := rows.Scan(&ig.ErrorType, &ig.Priority, &ig.Count)
		if err != nil {
			log.Println(err)
			continue
		}

		issues = append(issues, ig)
	}

	return issues
}

func (ds *Datastore) FindErrorTypesByPage(pid int, cid int64) []string {
	var et []string
	query := `
		SELECT 
			issue_types.type
		FROM issues
		INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
		WHERE pagereport_id = ? and crawl_id = ?
		GROUP BY issue_type_id`

	rows, err := ds.db.Query(query, pid, cid)
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

func (ds *Datastore) GetNumberOfPagesForIssues(cid int64, errorType string) int {
	query := `
		SELECT count(DISTINCT pagereport_id)
		FROM issues
		INNER JOIN issue_types ON issue_types.id = issues.issue_type_id
		WHERE issue_types.type = ? AND crawl_id  = ?`

	row := ds.db.QueryRow(query, errorType, cid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("GetNumberOfPagesForIssues: %v\n", err)
	}
	var f float64 = float64(c) / float64(paginationMax)
	return int(math.Ceil(f))
}
