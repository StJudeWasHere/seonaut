package repository

import (
	"database/sql"
	"log"

	"github.com/stjudewashere/seonaut/internal/models"
)

type ProjectRepository struct {
	DB *sql.DB
}

// SaveProject inserts a new project into the database.
func (ds *ProjectRepository) SaveProject(project *models.Project, uid int) {
	query := `
		INSERT INTO projects (
			url,
			ignore_robotstxt,
			follow_nofollow,
			include_noindex,
			crawl_sitemap,
			allow_subdomains,
			basic_auth,
			user_id,
			check_external_links,
			archive,
			user_agent
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, _ := ds.DB.Prepare(query)
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
		project.CheckExternalLinks,
		project.Archive,
		project.UserAgent,
	)
	if err != nil {
		log.Printf("saveProject: %v\n", err)
	}
}

// FindProjectsByUser returns a slice with all the projects of the specified user.
func (ds *ProjectRepository) FindProjectsByUser(uid int) []models.Project {
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
			created,
			check_external_links,
			archive,
			user_agent
		FROM projects
		WHERE user_id = ?
		ORDER BY url ASC`

	rows, err := ds.DB.Query(query, uid)
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
			&p.CheckExternalLinks,
			&p.Archive,
			&p.UserAgent,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		projects = append(projects, p)
	}

	return projects
}

// Returns a Project model with the speciefied id and user id.
func (ds *ProjectRepository) FindProjectById(id int, uid int) (models.Project, error) {
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
			created,
			check_external_links,
			archive,
			user_agent
		FROM projects
		WHERE id = ? AND user_id = ?`

	row := ds.DB.QueryRow(query, id, uid)

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
		&p.CheckExternalLinks,
		&p.Archive,
		&p.UserAgent,
	)
	if err != nil {
		log.Println(err)
		return p, err
	}

	return p, nil
}

// DisableProject disables a project marking it as "deleting".
func (ds *ProjectRepository) DisableProject(p *models.Project) {
	query := `UPDATE projects SET deleting=1 WHERE id = ?`
	_, err := ds.DB.Exec(query, p.Id)
	if err != nil {
		log.Printf("DeleteProject: update: pid %d %v\n", p.Id, err)
	}
}

// DeleteProject deletes the project.
func (ds *ProjectRepository) DeleteProject(p *models.Project) {
	query := `DELETE FROM projects WHERE id = ?`
	_, err := ds.DB.Exec(query, p.Id)
	if err != nil {
		log.Printf("DeleteProject: pid %d %v\n", p.Id, err)
		return
	}
}

// UpdateProject updates a project with the data specified in the Project model.
func (ds *ProjectRepository) UpdateProject(p *models.Project) error {
	query := `
		UPDATE projects SET
			ignore_robotstxt = ?,
			follow_nofollow = ?,
			include_noindex = ?,
			crawl_sitemap = ?,
			allow_subdomains = ?,
			basic_auth = ?,
			check_external_links = ?,
			archive = ?,
			user_agent = ?
		WHERE id = ?
	`
	_, err := ds.DB.Exec(
		query,
		p.IgnoreRobotsTxt,
		p.FollowNofollow,
		p.IncludeNoindex,
		p.CrawlSitemap,
		p.AllowSubdomains,
		p.BasicAuth,
		p.CheckExternalLinks,
		p.Archive,
		p.UserAgent,
		p.Id,
	)

	return err
}
