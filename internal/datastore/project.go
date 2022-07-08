package datastore

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/project"
)

func (ds *Datastore) SaveProject(project *project.Project, uid int) {
	query := `
		INSERT INTO projects (
			url,
			ignore_robotstxt,
			follow_nofollow,
			include_noindex,
			crawl_sitemap,
			allow_subdomains,
			user_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()
	_, err := stmt.Exec(
		project.URL,
		project.IgnoreRobotsTxt,
		project.FollowNofollow,
		project.IncludeNoindex,
		project.CrawlSitemap,
		project.AllowSubdomains,
		uid,
	)
	if err != nil {
		log.Printf("saveProject: %v\n", err)
	}
}

func (ds *Datastore) FindProjectsByUser(uid int) []project.Project {
	var projects []project.Project
	query := `
		SELECT
			id,
			url,
			ignore_robotstxt,
			follow_nofollow,
			include_noindex,
			crawl_sitemap,
			allow_subdomains,
			deleting,
			created
		FROM projects
		WHERE user_id = ?
		ORDER BY url ASC`

	rows, err := ds.db.Query(query, uid)
	if err != nil {
		log.Println(err)
		return projects
	}

	for rows.Next() {
		p := project.Project{}
		err := rows.Scan(
			&p.Id,
			&p.URL,
			&p.IgnoreRobotsTxt,
			&p.FollowNofollow,
			&p.IncludeNoindex,
			&p.CrawlSitemap,
			&p.AllowSubdomains,
			&p.Deleting,
			&p.Created,
		)
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
		SELECT
			id,
			url,
			ignore_robotstxt,
			follow_nofollow,
			include_noindex,
			crawl_sitemap,
			allow_subdomains,
			deleting,
			created
		FROM projects
		WHERE id = ? AND user_id = ?`

	row := ds.db.QueryRow(query, id, uid)

	p := project.Project{}
	err := row.Scan(
		&p.Id,
		&p.URL,
		&p.IgnoreRobotsTxt,
		&p.FollowNofollow,
		&p.IncludeNoindex,
		&p.CrawlSitemap,
		&p.AllowSubdomains,
		&p.Deleting,
		&p.Created,
	)
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
	query := `
		UPDATE
			crawls
		SET
			end = ?,
			total_urls = ?,
			blocked_by_robotstxt = ?,
			noindex = ?,
			robotstxt_exists = ?,
			sitemap_exists = ?
		WHERE id = ?
	`
	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	c.End = sql.NullTime{
		Time:  time.Now(),
		Valid: true,
	}

	_, err := stmt.Exec(
		c.End,
		c.TotalURLs,
		c.BlockedByRobotstxt,
		c.Noindex,
		c.RobotstxtExists,
		c.SitemapExists,
		c.Id,
	)
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
			issues_end,
			robotstxt_exists,
			sitemap_exists
		FROM crawls
		WHERE project_id = ?
		ORDER BY start DESC LIMIT 1`

	row := ds.db.QueryRow(query, p.Id)

	crawl := crawler.Crawl{}
	err := row.Scan(
		&crawl.Id,
		&crawl.Start,
		&crawl.End,
		&crawl.TotalURLs,
		&crawl.TotalIssues,
		&crawl.IssuesEnd,
		&crawl.RobotstxtExists,
		&crawl.SitemapExists,
	)
	if err != nil {
		log.Printf("GetLastCrawl: %v\n", err)
	}

	return crawl
}

func (ds *Datastore) GetLastCrawls(p project.Project, limit int) []crawler.Crawl {
	query := `
		SELECT
			id,
			start,
			end,
			total_urls,
			total_issues,
			issues_end,
			critical_issues,
			warning_issues,
			notice_issues,
			blocked_by_robotstxt,
			noindex
		FROM crawls
		WHERE project_id = ?
		ORDER BY start DESC LIMIT ?`

	crawls := []crawler.Crawl{}
	rows, err := ds.db.Query(query, p.Id, limit)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		crawl := crawler.Crawl{}
		err := rows.Scan(
			&crawl.Id,
			&crawl.Start,
			&crawl.End,
			&crawl.TotalURLs,
			&crawl.TotalIssues,
			&crawl.IssuesEnd,
			&crawl.CriticalIssues,
			&crawl.WarningIssues,
			&crawl.NoticeIssues,
			&crawl.BlockedByRobotstxt,
			&crawl.Noindex,
		)
		if err != nil {
			log.Printf("GetLastCrawl: %v\n", err)
		}
		crawls = append([]crawler.Crawl{crawl}, crawls...)
	}

	return crawls
}

func (ds *Datastore) SaveIssuesCount(crawlId int64, critical, warning, notice int) {
	query := `UPDATE
		crawls
		SET critical_issues = ?, warning_issues = ?, notice_issues = ?
		WHERE id = ?`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()
	_, err := stmt.Exec(critical, warning, notice, crawlId)
	if err != nil {
		log.Printf("SaveIssuesCount: %v\n", err)
	}
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

func (ds *Datastore) DeleteProject(p *project.Project) {
	query := `UPDATE projects SET deleting=1 WHERE id = ?`
	_, err := ds.db.Exec(query, p.Id)
	if err != nil {
		log.Printf("DeleteProject: update: pid %d %v\n", p.Id, err)
		return
	}

	go func() {
		c := ds.GetLastCrawl(p)

		ds.DeleteCrawl(c.Id)

		query := `DELETE FROM projects WHERE id = ?`
		_, err := ds.db.Exec(query, p.Id)
		if err != nil {
			log.Printf("DeleteProject: pid %d %v\n", p.Id, err)
			return
		}
	}()
}

func (ds *Datastore) DeletePreviousCrawl(pid int) {
	previousCrawl := ds.FindPreviousCrawlId(pid)

	ds.DeleteCrawl(int64(previousCrawl))
}

func (ds *Datastore) DeleteCrawl(cid int64) {
	var deleteFunc func(cid int64, table string)
	deleteFunc = func(cid int64, table string) {
		query := fmt.Sprintf("DELETE FROM %s WHERE crawl_id = ? ORDER BY id DESC LIMIT 1000", table)
		_, err := ds.db.Exec(query, cid)
		if err != nil {
			log.Printf("DeleteCrawl: cid %d table %s %v\n", cid, table, err)
			return
		}

		query = fmt.Sprintf("SELECT count(*) FROM %s WHERE crawl_id = ?", table)
		row := ds.db.QueryRow(query, cid)
		var c int
		if err := row.Scan(&c); err != nil {
			log.Printf("DeleteCrawl count: pid %d table %s %v\n", cid, table, err)
		}

		if c > 0 {
			time.Sleep(1500 * time.Millisecond)
			deleteFunc(cid, table)
		}
	}

	deleteFunc(cid, "links")
	deleteFunc(cid, "external_links")
	deleteFunc(cid, "hreflangs")
	deleteFunc(cid, "issues")
	deleteFunc(cid, "images")
	deleteFunc(cid, "scripts")
	deleteFunc(cid, "styles")
	deleteFunc(cid, "iframes")
	deleteFunc(cid, "audios")
	deleteFunc(cid, "videos")
	deleteFunc(cid, "pagereports")
}
