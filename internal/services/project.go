package services

import (
	"errors"
	"net/url"
	"strings"

	"github.com/stjudewashere/seonaut/internal/models"
)

type (
	ProjectServiceRepository interface {
		SaveProject(*models.Project, int)
		DeleteProject(*models.Project)
		DisableProject(*models.Project)
		UpdateProject(p *models.Project) error
		FindProjectById(id int, uid int) (models.Project, error)

		DeleteProjectCrawls(*models.Project)
	}

	ProjectService struct {
		repository ProjectServiceRepository
	}
)

func NewProjectService(r ProjectServiceRepository) *ProjectService {
	return &ProjectService{repository: r}
}

// SaveProject stores a new project.
// It trims the spaces in the project's URL field and checks the scheme to
// make sure it is http or https.
func (s *ProjectService) SaveProject(project *models.Project, userId int) error {
	project.URL = strings.TrimSpace(project.URL)
	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		return err
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("protocol not supported")
	}

	s.repository.SaveProject(project, userId)

	return nil
}

// Return a project specified by id and user.
// It populates the Host field from the project's URL.
func (s *ProjectService) FindProject(id, uid int) (models.Project, error) {
	project, err := s.repository.FindProjectById(id, uid)
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

// Delete a project and its related data.
func (s *ProjectService) DeleteProject(p *models.Project) {
	s.repository.DisableProject(p)
	go func() {
		s.repository.DeleteProjectCrawls(p)
		s.repository.DeleteProject(p)
	}()
}

// Update project details.
func (s *ProjectService) UpdateProject(p *models.Project) error {
	return s.repository.UpdateProject(p)
}
