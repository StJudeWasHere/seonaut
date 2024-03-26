package container

import (
	"net/url"

	"github.com/stjudewashere/seonaut/internal/models"
)

type ProjectViewStorage interface {
	FindProjectsByUser(int) []models.Project
	FindProjectById(id int, uid int) (models.Project, error)
	GetLastCrawl(*models.Project) models.Crawl
}

type ProjectViewService struct {
	storage ProjectViewStorage
}

type ProjectView struct {
	Project models.Project
	Crawl   models.Crawl
}

func NewProjectViewService(s ProjectViewStorage) *ProjectViewService {
	return &ProjectViewService{
		storage: s,
	}
}

// GetProjectView returns a new ProjectView with the specified project
// and the project's last crawl.
func (s *ProjectViewService) GetProjectView(id, uid int) (*ProjectView, error) {
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

	v := &ProjectView{
		Project: project,
		Crawl:   c,
	}

	return v, nil
}

// GetProjectViews returns a slice of ProjectViews with all of the user's
// projects and its last crawls.
func (s *ProjectViewService) GetProjectViews(uid int) []ProjectView {
	var views []ProjectView

	projects := s.storage.FindProjectsByUser(uid)
	for _, p := range projects {
		pv := ProjectView{
			Project: p,
			Crawl:   s.storage.GetLastCrawl(&p),
		}
		views = append(views, pv)
	}

	return views
}
