package datastore

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/stjudewashere/seonaut/internal/models"
)

func (ds *Datastore) SaveProject(project *models.Project, uid int) {
	query := `
		INSERT INTO projects (
			url,
			ignore_robotstxt,
			follow_nofollow,
			include_noindex,
			crawl_sitemap,
			allow_subdomains,
			basic_auth,
			user_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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
		project.BasicAuth,
		uid,
	)
	if err != nil {
		log.Printf("saveProject: %v\n", err)
	}
}

func (ds *Datastore) FindProjectsByUser(uid int) []models.Project {
	var projects []models.Project
	query := `
		SELECT
			id,
			url,
			ignore_robotstxt,
			follow_nofollow,
			include_noindex,
			crawl_sitemap,
			allow_subdomains,
			basic_auth,
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
		p := models.Project{}
		err := rows.Scan(
			&p.Id,
			&p.URL,
			&p.IgnoreRobotsTxt,
			&p.FollowNofollow,
			&p.IncludeNoindex,
			&p.CrawlSitemap,
			&p.AllowSubdomains,
			&p.BasicAuth,
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

func (ds *Datastore) FindProjectById(id int, uid int) (models.Project, error) {
	query := `
		SELECT
			id,
			url,
			ignore_robotstxt,
			follow_nofollow,
			include_noindex,
			crawl_sitemap,
			allow_subdomains,
			basic_auth,
			deleting,
			created
		FROM projects
		WHERE id = ? AND user_id = ?`

	row := ds.db.QueryRow(query, id, uid)

	p := models.Project{}
	err := row.Scan(
		&p.Id,
		&p.URL,
		&p.IgnoreRobotsTxt,
		&p.FollowNofollow,
		&p.IncludeNoindex,
		&p.CrawlSitemap,
		&p.AllowSubdomains,
		&p.BasicAuth,
		&p.Deleting,
		&p.Created,
	)
	if err != nil {
		log.Println(err)
		return p, err
	}

	return p, nil
}

func (ds *Datastore) SaveCrawl(p models.Project) (*models.Crawl, error) {
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

	return &models.Crawl{
		Id:        cid,
		ProjectId: p.Id,
		URL:       p.URL,
		Start:     time.Now(),
	}, nil
}

func (ds *Datastore) SaveEndCrawl(c *models.Crawl) (*models.Crawl, error) {
	query := `
		UPDATE
			crawls
		SET
			end = ?,
			total_urls = ?,
			blocked_by_robotstxt = ?,
			noindex = ?,
			robotstxt_exists = ?,
			sitemap_exists = ?,
			sitemap_blocked = ?,
			links_internal_follow = ?,
			links_internal_nofollow = ?,
			links_external_follow = ?,
			links_external_nofollow = ?,
			links_sponsored = ?,
			links_ugc = ?
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
		c.SitemapIsBlocked,
		c.InternalFollowLinks,
		c.InternalNoFollowLinks,
		c.ExternalFollowLinks,
		c.ExternalNoFollowLinks,
		c.SponsoredLinks,
		c.UGCLinks,
		c.Id,
	)
	if err != nil {
		log.Printf("saveEndCrawl: %v\n", err)
		return c, err
	}

	return c, nil
}

func (ds *Datastore) GetLastCrawl(p *models.Project) models.Crawl {
	query := `
		SELECT
			id,
			start,
			end,
			total_urls,
			total_issues,
			critical_issues,
			alert_issues,
			warning_issues,
			issues_end,
			robotstxt_exists,
			sitemap_exists,
			sitemap_blocked,
			links_internal_follow,
			links_internal_nofollow,
			links_external_follow,
			links_external_nofollow,
			links_sponsored,
			links_ugc
		FROM crawls
		WHERE project_id = ?
		ORDER BY start DESC LIMIT 1`

	row := ds.db.QueryRow(query, p.Id)

	crawl := models.Crawl{}
	err := row.Scan(
		&crawl.Id,
		&crawl.Start,
		&crawl.End,
		&crawl.TotalURLs,
		&crawl.TotalIssues,
		&crawl.CriticalIssues,
		&crawl.AlertIssues,
		&crawl.WarningIssues,
		&crawl.IssuesEnd,
		&crawl.RobotstxtExists,
		&crawl.SitemapExists,
		&crawl.SitemapIsBlocked,
		&crawl.InternalFollowLinks,
		&crawl.InternalNoFollowLinks,
		&crawl.ExternalFollowLinks,
		&crawl.ExternalNoFollowLinks,
		&crawl.SponsoredLinks,
		&crawl.UGCLinks,
	)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("GetLastCrawl project id %d: %v\n", p.Id, err)
	}

	return crawl
}

func (ds *Datastore) GetLastCrawls(p models.Project, limit int) []models.Crawl {
	query := `
		SELECT
			id,
			start,
			end,
			total_urls,
			total_issues,
			issues_end,
			critical_issues,
			alert_issues,
			warning_issues,
			blocked_by_robotstxt,
			noindex
		FROM crawls
		WHERE project_id = ?
		ORDER BY start DESC LIMIT ?`

	crawls := []models.Crawl{}
	rows, err := ds.db.Query(query, p.Id, limit)
	if err != nil {
		log.Println(err)
	}

	for rows.Next() {
		crawl := models.Crawl{}
		err := rows.Scan(
			&crawl.Id,
			&crawl.Start,
			&crawl.End,
			&crawl.TotalURLs,
			&crawl.TotalIssues,
			&crawl.IssuesEnd,
			&crawl.CriticalIssues,
			&crawl.AlertIssues,
			&crawl.WarningIssues,
			&crawl.BlockedByRobotstxt,
			&crawl.Noindex,
		)
		if err != nil {
			log.Printf("GetLastCrawl: %v\n", err)
		}
		crawls = append([]models.Crawl{crawl}, crawls...)
	}

	return crawls
}

func (ds *Datastore) SaveIssuesCount(crawlId int64, critical, alert, warning int) {
	query := `UPDATE
		crawls
		SET 
			critical_issues = ?,
			alert_issues = ?,
			warning_issues = ?,
			total_issues = ?
		WHERE id = ?`

	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	total := critical + alert + warning
	_, err := stmt.Exec(critical, alert, warning, total, crawlId)
	if err != nil {
		log.Printf("SaveIssuesCount: %v\n", err)
	}
}

func (ds *Datastore) DeleteProject(p *models.Project) {
	query := `UPDATE projects SET deleting=1 WHERE id = ?`
	_, err := ds.db.Exec(query, p.Id)
	if err != nil {
		log.Printf("DeleteProject: update: pid %d %v\n", p.Id, err)
		return
	}

	go func(p models.Project) {
		ds.DeleteProjectCrawls(&p)

		query := `DELETE FROM projects WHERE id = ?`
		_, err := ds.db.Exec(query, p.Id)
		if err != nil {
			log.Printf("DeleteProject: pid %d %v\n", p.Id, err)
			return
		}
	}(*p)
}

func (ds *Datastore) UpdateProject(p *models.Project) error {
	query := `
		UPDATE projects SET
			ignore_robotstxt = ?,
			follow_nofollow = ?,
			include_noindex = ?,
			crawl_sitemap = ?,
			allow_subdomains = ?,
			basic_auth = ?
		WHERE id = ?
	`
	_, err := ds.db.Exec(
		query,
		p.IgnoreRobotsTxt,
		p.FollowNofollow,
		p.IncludeNoindex,
		p.CrawlSitemap,
		p.AllowSubdomains,
		p.BasicAuth,
		p.Id,
	)
	if err != nil {
		log.Printf("UpdateProject: pid %d %v\n", p.Id, err)
	}

	return err
}

func (ds *Datastore) GetPreviousCrawl(p *models.Project) (*models.Crawl, error) {
	query := `
		SELECT
			id,
			start,
			end,
			total_urls,
			total_issues,
			issues_end,
			critical_issues,
			alert_issues,
			warning_issues,
			blocked_by_robotstxt,
			noindex
		FROM crawls
		WHERE project_id = ?
		ORDER BY end DESC
		LIMIT 1, 1`

	row := ds.db.QueryRow(query, p.Id)

	crawl := &models.Crawl{}
	err := row.Scan(
		&crawl.Id,
		&crawl.Start,
		&crawl.End,
		&crawl.TotalURLs,
		&crawl.TotalIssues,
		&crawl.IssuesEnd,
		&crawl.CriticalIssues,
		&crawl.AlertIssues,
		&crawl.WarningIssues,
		&crawl.BlockedByRobotstxt,
		&crawl.Noindex,
	)

	return crawl, err
}

func (ds *Datastore) DeleteCrawlData(crawl *models.Crawl) {
	var deleteFunc func(cid int64, table string)
	deleteFunc = func(cid int64, table string) {
		query := fmt.Sprintf("DELETE FROM %s WHERE crawl_id = ? ORDER BY id DESC LIMIT 1000", table)
		_, err := ds.db.Exec(query, cid)
		if err != nil {
			log.Printf("DeleteCrawlData: cid %d table %s %v\n", cid, table, err)
			return
		}

		query = fmt.Sprintf("SELECT count(*) FROM %s WHERE crawl_id = ?", table)
		row := ds.db.QueryRow(query, cid)
		var c int
		if err := row.Scan(&c); err != nil {
			log.Printf("DeleteCrawlData count: pid %d table %s %v\n", cid, table, err)
		}

		if c > 0 {
			time.Sleep(1500 * time.Millisecond)
			deleteFunc(cid, table)
		}
	}

	deleteFunc(crawl.Id, "links")
	deleteFunc(crawl.Id, "external_links")
	deleteFunc(crawl.Id, "hreflangs")
	deleteFunc(crawl.Id, "issues")
	deleteFunc(crawl.Id, "images")
	deleteFunc(crawl.Id, "scripts")
	deleteFunc(crawl.Id, "styles")
	deleteFunc(crawl.Id, "iframes")
	deleteFunc(crawl.Id, "audios")
	deleteFunc(crawl.Id, "videos")
	deleteFunc(crawl.Id, "pagereports")
}

// DeleteProjectCrawls deletes the project's crawl data
func (ds *Datastore) DeleteProjectCrawls(p *models.Project) {
	query := `
		SELECT
			id
		FROM crawls
		WHERE project_id = ?
	`

	rows, err := ds.db.Query(query, p.Id)
	if err != nil {
		log.Printf("DeleteProjectCrawls Query: %v\n", err)
	}

	for rows.Next() {
		c := &models.Crawl{}
		if err := rows.Scan(&c.Id); err != nil {
			log.Printf("DeleteProjectCrawls: %v\n", err)
		}

		ds.DeleteCrawlData(c)
	}

	query = `DELETE FROM crawls WHERE project_id = ?`
	_, err = ds.db.Exec(query, p.Id)
	if err != nil {
		log.Printf("deleting crawls for project %d: %v", p.Id, err)
		return
	}
}

// Deletes all crawls that are unfinished and have the issues_end field set to null.
// It cleans up the crawl data for each unfinished crawl before deleting it.
func (ds *Datastore) DeleteUnfinishedCrawls() int {
	query := `
		SELECT
			crawls.id
		FROM crawls
		WHERE crawls.issues_end IS NULL
	`
	count := 0

	rows, err := ds.db.Query(query)
	if err != nil {
		log.Println(err)
		return count
	}

	ids := []any{}
	placeholders := []string{}
	for rows.Next() {
		c := &models.Crawl{}
		err := rows.Scan(&c.Id)
		if err != nil {
			log.Printf("DeleteUnfinishedCrawls: %v\n", err)
			continue
		}

		count++
		ds.DeleteCrawlData(c)
		ids = append(ids, c.Id)
		placeholders = append(placeholders, "?")
	}

	if len(ids) == 0 {
		return count
	}

	placeholdersStr := strings.Join(placeholders, ",")
	deleteQuery := fmt.Sprintf("DELETE FROM crawls WHERE id IN (%s)", placeholdersStr)
	_, err = ds.db.Exec(deleteQuery, ids...)
	if err != nil {
		log.Printf("DeleteUnfinishedCrawls: %v", err)
	}

	return count
}
