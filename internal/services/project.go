package services

import (
	"errors"
	"net/url"

	"github.com/stjudewashere/seonaut/internal/models"
)

type (
	ProjectServiceStorage interface {
		SaveProject(*models.Project, int)
		DeleteProject(*models.Project)
		GetLastCrawl(*models.Project) models.Crawl
		UpdateProject(p *models.Project) error
		FindProjectById(id int, uid int) (models.Project, error)
	}

	ProjectService struct {
		storage      ProjectServiceStorage
		cacheManager *CacheManager
	}
)

func NewProjectService(s ProjectServiceStorage, cm *CacheManager) *ProjectService {
	return &ProjectService{
		storage:      s,
		cacheManager: cm,
	}
}

// SaveProject stores a new project.
func (s *ProjectService) SaveProject(project *models.Project, userId int) error {
	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		return err
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("protocol not supported")
	}

	s.storage.SaveProject(project, userId)

	return nil
}

// Return a project specified by id and user.
// It populates the Host field from the project's URL.
func (s *ProjectService) FindProject(id, uid int) (models.Project, error) {
	project, err := s.storage.FindProjectById(id, uid)
	if err != nil {
		return project, err
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		return project, err
	}

	project.Host = parsedURL.Host

	return project, nil
}

// Delete a project and remove any related data that has been cached.
func (s *ProjectService) DeleteProject(p *models.Project) {
	last := s.storage.GetLastCrawl(p)
	s.cacheManager.RemoveCrawlCache(&last)
	s.storage.DeleteProject(p)
}

// Update project details.
func (s *ProjectService) UpdateProject(p *models.Project) error {
	return s.storage.UpdateProject(p)
}
