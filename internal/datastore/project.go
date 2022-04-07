package datastore

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/project"
)

func (ds *Datastore) SaveProject(s string, ignoreRobotsTxt, followNofollow, includeNoindex bool, uid int) {
	query := `
		INSERT INTO projects (url, ignore_robotstxt, follow_nofollow, include_noindex, user_id)
		VALUES (?, ?, ?, ?, ?)
	`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()
	_, err := stmt.Exec(s, ignoreRobotsTxt, followNofollow, includeNoindex, uid)
	if err != nil {
		log.Printf("saveProject: %v\n", err)
	}
}

func (ds *Datastore) FindProjectsByUser(uid int) []project.Project {
	var projects []project.Project
	query := `
		SELECT id, url, ignore_robotstxt, follow_nofollow, include_noindex, created
		FROM projects
		WHERE user_id = ?`

	rows, err := ds.db.Query(query, uid)
	if err != nil {
		log.Println(err)
		return projects
	}

	for rows.Next() {
		p := project.Project{}
		err := rows.Scan(&p.Id, &p.URL, &p.IgnoreRobotsTxt, &p.FollowNofollow, &p.IncludeNoindex, &p.Created)
		if err != nil {
			log.Println(err)
			continue
		}

		projects = append(projects, p)
	}

	return projects
}

func (ds *Datastore) FindProjectById(id int, uid int) (project.Project, error) {
	query := `
		SELECT id, url, ignore_robotstxt, follow_nofollow, include_noindex, created
		FROM projects
		WHERE id = ? AND user_id = ?`

	row := ds.db.QueryRow(query, id, uid)

	p := project.Project{}
	err := row.Scan(&p.Id, &p.URL, &p.IgnoreRobotsTxt, &p.FollowNofollow, &p.IncludeNoindex, &p.Created)
	if err != nil {
		log.Println(err)
		return p, err
	}

	return p, nil
}

func (ds *Datastore) SaveCrawl(p project.Project) (*crawler.Crawl, error) {
	stmt, _ := ds.db.Prepare("INSERT INTO crawls (project_id) VALUES (?)")
	defer stmt.Close()
	res, err := stmt.Exec(p.Id)

	if err != nil {
		return nil, err
	}

	cid, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &crawler.Crawl{
		Id:        cid,
		ProjectId: p.Id,
		URL:       p.URL,
		Start:     time.Now(),
	}, nil
}

func (ds *Datastore) SaveEndCrawl(c *crawler.Crawl) (*crawler.Crawl, error) {
	stmt, _ := ds.db.Prepare("UPDATE crawls SET end = ?, total_urls= ? WHERE id = ?")
	defer stmt.Close()

	c.End = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	_, err := stmt.Exec(c.End, c.TotalURLs, c.Id)
	if err != nil {
		log.Printf("saveEndCrawl: %v\n", err)
		return c, err
	}

	return c, nil
}

func (ds *Datastore) GetLastCrawl(p *project.Project) crawler.Crawl {
	query := `
		SELECT
			id,
			start,
			end,
			total_urls,
			total_issues,
			issues_end
		FROM crawls
		WHERE project_id = ?
		ORDER BY start DESC LIMIT 1`

	row := ds.db.QueryRow(query, p.Id)

	crawl := crawler.Crawl{}
	err := row.Scan(&crawl.Id, &crawl.Start, &crawl.End, &crawl.TotalURLs, &crawl.TotalIssues, &crawl.IssuesEnd)
	if err != nil {
		log.Printf("GetLastCrawl: %v\n", err)
	}

	return crawl
}

func (ds *Datastore) FindPreviousCrawlId(pid int) int {
	query := `
		SELECT
			id
		FROM crawls
		WHERE project_id = ?
		ORDER BY end DESC
		LIMIT 1, 1`

	row := ds.db.QueryRow(query, pid)
	var c int
	if err := row.Scan(&c); err != nil {
		log.Printf("FindPreviousCrawlId: %v\n", err)
	}

	return c
}

func (ds *Datastore) DeletePreviousCrawl(pid int) {
	previousCrawl := ds.FindPreviousCrawlId(pid)

	var deleteFunc func(cid int, table string)
	deleteFunc = func(cid int, table string) {
		query := fmt.Sprintf("DELETE FROM %s WHERE crawl_id = ? ORDER BY id DESC LIMIT 1000", table)
		_, err := ds.db.Exec(query, previousCrawl)
		if err != nil {
			log.Printf("DeletePreviousCeawl: pid %d table %s %v\n", pid, table, err)
			return
		}

		query = fmt.Sprintf("SELECT count(*) FROM %s WHERE crawl_id = ?", table)
		row := ds.db.QueryRow(query, previousCrawl)
		var c int
		if err := row.Scan(&c); err != nil {
			log.Printf("DeletePreviousCrawl count: pid %d table %s %v\n", pid, table, err)
		}

		if c > 0 {
			time.Sleep(1500 * time.Millisecond)
			deleteFunc(cid, table)
		}
	}

	deleteFunc(previousCrawl, "links")
	deleteFunc(previousCrawl, "external_links")
	deleteFunc(previousCrawl, "hreflangs")
	deleteFunc(previousCrawl, "issues")
	deleteFunc(previousCrawl, "images")
	deleteFunc(previousCrawl, "scripts")
	deleteFunc(previousCrawl, "styles")
	deleteFunc(previousCrawl, "pagereports")
}
