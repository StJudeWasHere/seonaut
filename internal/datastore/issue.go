package datastore

import (
	"log"
	"math"
	"sort"
	"time"

	"github.com/stjudewashere/seonaut/internal/issue"
)

func (ds *Datastore) CountByMediaType(cid int64) issue.CountList {
	query := `
		SELECT media_type, count(*)
		FROM pagereports
		WHERE crawl_id = ? AND crawled = 1
		GROUP BY media_type`

	return ds.countListQuery(query, cid)
}

func (ds *Datastore) CountByStatusCode(cid int64) issue.CountList {
	query := `
		SELECT
			status_code,
			count(*)
		FROM pagereports
		WHERE crawl_id = ? AND crawled = 1
		GROUP BY status_code`

	return ds.countListQuery(query, cid)
}

func (ds *Datastore) CountByFollowLinks(cid int64) issue.CountList {
	query := `
		SELECT
			IF(nofollow, "nofollow", "follow"),
			count(*)
		FROM links
		WHERE crawl_id = ?
		GROUP BY nofollow
		ORDER BY nofollow ASC`

	return ds.countListQuery(query, cid)
}

func (ds *Datastore) CountByFollowExternalLinks(cid int64) issue.CountList {
	query := `
		SELECT
			IF(nofollow, "nofollow", "follow"),
			count(*)
		FROM external_links
		WHERE crawl_id = ?
		GROUP BY nofollow
		ORDER BY nofollow ASC`

	return ds.countListQuery(query, cid)
}

func (ds *Datastore) countListQuery(query string, cid int64) issue.CountList {
	m := issue.CountList{}
	rows, err := ds.db.Query(query, cid)
	if err != nil {
		log.Println(err)
		return m
	}

	for rows.Next() {
		c := issue.CountItem{}
		err := rows.Scan(&c.Key, &c.Value)
		if err != nil {
			log.Println(err)
			continue
		}
		m = append(m, c)
	}

	sort.Sort(sort.Reverse(m))

	return m
}

func (ds *Datastore) SaveEndIssues(cid int64, t time.Time, totalIssues int) {
	stmt, _ := ds.db.Prepare("UPDATE crawls SET issues_end = ?, total_issues = ? WHERE id = ?")
	defer stmt.Close()
	_, err := stmt.Exec(t, totalIssues, cid)
	if err != nil {
		log.Printf("saveEndIssues: %v\n", err)
	}
}

func (ds *Datastore) SaveIssues(iStream <-chan *issue.Issue) {
	query := `
		INSERT INTO issues (pagereport_id, crawl_id, issue_type_id)
		VALUES (?, ?, ?)`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	for i := range iStream {
		_, err := stmt.Exec(i.PageReportId, i.CrawlId, i.ErrorType)
		if err != nil {
			log.Printf("saveIssues -> ID: %d ERROR: %d CRAWL: %d %v\n", i.PageReportId, i.ErrorType, i.CrawlId, err)
			continue
		}
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
