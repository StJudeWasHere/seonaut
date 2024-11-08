package services

import (
	"errors"
	"net/url"
	"os"
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
		FindProjectsByUser(userId int) []models.Project

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
		s.DeleteArchive(p)
	}()
}

// Update project details.
func (s *ProjectService) UpdateProject(p *models.Project) error {
	if !p.Archive {
		s.DeleteArchive(p)
	}

	return s.repository.UpdateProject(p)
}

// Delete all user projects and crawl data.
func (s *ProjectService) DeleteAllUserProjects(user *models.User) {
	projects := s.repository.FindProjectsByUser(user.Id)
	for _, p := range projects {
		s.repository.DeleteProjectCrawls(&p)
		s.repository.DeleteProject(&p)
		s.DeleteArchive(&p)
	}
}

// ArchiveExists checks if a wacz file exists for the current project.
// It returns true if it exists, otherwise it returns false.
func (s *ProjectService) ArchiveExists(p *models.Project) bool {
	_, err := os.Stat(ArchiveDir + p.Host + ".wacz")
	return err == nil
}

// DeleteArchive removes the wacz archive file for a given project.
// It checks if the file exists before removing it.
func (s *ProjectService) DeleteArchive(p *models.Project) {
	if !s.ArchiveExists(p) {
		return
	}

	os.Remove(ArchiveDir + p.Host + ".wacz")
}

// GetArchiveFilePath returns the project's wacz file path if it exists,
// otherwise it returns an error.
func (s *ProjectService) GetArchiveFilePath(p *models.Project) (string, error) {
	if !s.ArchiveExists(p) {
		return "", errors.New("WACZ archive file does not exist")
	}

	return ArchiveDir + p.Host + ".wacz", nil
}
