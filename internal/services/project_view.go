package services

import (
	"net/url"

	"github.com/stjudewashere/seonaut/internal/models"
)

type (
	ProjectViewServiceStorage interface {
		FindProjectsByUser(int) []models.Project
		FindProjectById(id int, uid int) (models.Project, error)

		GetLastCrawl(*models.Project) models.Crawl
	}

	ProjectViewService struct {
		storage ProjectViewServiceStorage
	}
)

func NewProjectViewService(s ProjectViewServiceStorage) *ProjectViewService {
	return &ProjectViewService{storage: s}
}

// GetProjectView returns a new ProjectView with the specified project
// and the project's last crawl.
func (s *ProjectViewService) GetProjectView(id, uid int) (*models.ProjectView, error) {
	project, err := s.storage.FindProjectById(id, uid)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		return nil, err
	}

	project.Host = parsedURL.Host

	c := s.storage.GetLastCrawl(&project)

	v := &models.ProjectView{
		Project: project,
		Crawl:   c,
	}

	return v, nil
}

// GetProjectViews returns a slice of ProjectViews with all of the user's
// projects and its last crawls.
func (s *ProjectViewService) GetProjectViews(uid int) []models.ProjectView {
	var views []models.ProjectView

	projects := s.storage.FindProjectsByUser(uid)
	for _, p := range projects {
		pv := models.ProjectView{
			Project: p,
			Crawl:   s.storage.GetLastCrawl(&p),
		}
		views = append(views, pv)
	}

	return views
}
