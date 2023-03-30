package project_test

import (
	"errors"
	"testing"

	"github.com/stjudewashere/seonaut/internal/cache_manager"
	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/project"
)

const (
	gid        = 1
	guid       = 1
	projectURL = "https://example.com"
	urlHost    = "example.com"
	urlScheme  = "https"
)

type storage struct{}

func (s *storage) SaveProject(project *models.Project, userId int) {}
func (s *storage) DeleteProject(project *models.Project)           {}
func (s *storage) UpdateProject(p *models.Project) error {
	return nil
}
func (s *storage) GetLastCrawl(p *models.Project) models.Crawl {
	return models.Crawl{}
}
func (s *storage) FindProjectById(id, uid int) (models.Project, error) {
	p := models.Project{}

	if id != gid || uid != guid {
		return p, errors.New("Project does not exist")
	}

	p.URL = projectURL

	return p, nil
}

var service = project.NewService(&storage{}, cache_manager.New())

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

func TestSaveProject(t *testing.T) {
	// Valid URL
	err := service.SaveProject(&models.Project{URL: projectURL}, guid)
	if err != nil {
		t.Error("TestSaveProject: should not return error")
	}

	// Not valid URL
	err = service.SaveProject(&models.Project{URL: "...."}, guid)
	if err == nil {
		t.Error("TestSaveProject: invalid URL should return error")
	}

	// Not supported scheme
	err = service.SaveProject(&models.Project{URL: "ftp://example.org"}, guid)
	if err == nil {
		t.Error("TestSaveProject: not supported scheme should return error")
	}
}
