package services_test

import (
	"errors"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

const (
	gid        = 1
	guid       = 1
	projectURL = "https://example.com"
	userAgent  = "TEST UserAgent"
	urlHost    = "example.com"
	urlScheme  = "https"
)

// Create a test repository for the service.
type projectTestRepository struct{}

func (s *projectTestRepository) SaveProject(project *models.Project, userId int) {}
func (s *projectTestRepository) DeleteProject(project *models.Project)           {}
func (s *projectTestRepository) DisableProject(project *models.Project)          {}
func (s *projectTestRepository) UpdateProject(p *models.Project) error           { return nil }
func (p *projectTestRepository) FindProjectsByUser(uid int) []models.Project {
	return []models.Project{}
}
func (s *projectTestRepository) FindProjectById(id, uid int) (models.Project, error) {
	p := models.Project{}

	if id != gid || uid != guid {
		return p, errors.New("Project does not exist")
	}

	p.URL = projectURL

	return p, nil
}
func (s *projectTestRepository) DeleteProjectCrawls(*models.Project) {}

// Create an Archive Deleter for the service.
type ArchiveDeleter struct{}

func (ad *ArchiveDeleter) DeleteArchive(p *models.Project) {}

// Create the service with the test repository and archive deleter.
var service = services.NewProjectService(&projectTestRepository{}, &ArchiveDeleter{})

// Test FindProjectById. This function parses the URL to populate
// the Host field in the project model.
func TestFindProjectById(t *testing.T) {
	p, err := service.FindProject(gid, guid)
	if err != nil {
		t.Error(err)
	}

	if p.URL != projectURL {
		t.Errorf("p.URL: %s != %s", p.URL, projectURL)
	}

	if p.Host != urlHost {
		t.Errorf("p.Host: %s != %s", p.Host, urlHost)
	}

	p, err = service.FindProject(0, 0)
	if err == nil {
		t.Error("TestFindProjectById: should return err")
	}
}

// Test SaveProject with different project configurations.
func TestSaveProject(t *testing.T) {
	table := []struct {
		name      string
		project   *models.Project
		wantError bool
	}{
		{
			name:      "Valid URL and valid User-Agent",
			project:   &models.Project{URL: projectURL, UserAgent: userAgent},
			wantError: false,
		},
		{
			name:      "Valid URL and empty User-Agent",
			project:   &models.Project{URL: projectURL},
			wantError: true,
		},
		{
			name:      "Not valid URL",
			project:   &models.Project{URL: "....", UserAgent: userAgent},
			wantError: true,
		},
		{
			name:      "Not supported scheme",
			project:   &models.Project{URL: "ftp://example.org", UserAgent: userAgent},
			wantError: true,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			err := service.SaveProject(tt.project, guid)
			if (err != nil) != tt.wantError {
				t.Errorf("SaveProject() want error %v got %v", tt.wantError, err)
			}
		})
	}
}

// Test UpdateProject with different project configurations.
func TestUpdateProject(t *testing.T) {
	table := []struct {
		name      string
		project   *models.Project
		wantError bool
	}{
		{
			name:      "Valid User-Agent",
			project:   &models.Project{URL: projectURL, UserAgent: userAgent},
			wantError: false,
		},
		{
			name:      "Empty User-Agent",
			project:   &models.Project{URL: projectURL},
			wantError: true,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateProject(tt.project)
			if (err != nil) != tt.wantError {
				t.Errorf("SaveProject() want error %v got %v", tt.wantError, err)
			}
		})
	}
}
