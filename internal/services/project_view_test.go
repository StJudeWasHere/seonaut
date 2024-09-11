package services_test

import (
	"errors"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
)

const (
	test_uid          = 1
	test_pid          = 1
	test_cid          = 1
	test_url          = "https://example.org"
	test_total_models = 2
)

// Create a mock repository for the projectview service.
type projectViewTestRepository struct{}

func (s *projectViewTestRepository) FindProjectsByUser(uid int) []models.Project {
	m := []models.Project{}

	if uid != test_uid {
		return m
	}

	for i := 0; i < test_total_models; i++ {
		m = append(m, models.Project{})
	}

	return m
}

func (s *projectViewTestRepository) FindProjectById(id int, uid int) (models.Project, error) {
	if id == test_pid && uid == test_uid {
		return models.Project{Id: test_pid, URL: test_url}, nil
	}

	return models.Project{}, errors.New("Test error")
}

func (s *projectViewTestRepository) GetLastCrawl(p *models.Project) models.Crawl {
	if p.Id == test_pid {
		return models.Crawl{Id: test_cid}
	}

	return models.Crawl{}
}

var projectviewService = services.NewProjectViewService(&projectViewTestRepository{})

// TestGetProjectView tests the GetProjectView function of the projectview service.
// It verifies the behavior of the GetProjectView function with an existing projectview.
func TestGetProjectView_ProjectViewExists(t *testing.T) {
	pv, err := projectviewService.GetProjectView(test_pid, test_uid)
	if err != nil {
		t.Errorf("%v", err)

		return
	}

	if pv.Project.Id != test_pid {
		t.Errorf("Project Id != %d", test_pid)
	}

	if pv.Crawl.Id != test_cid {
		t.Errorf("Crawl Id != %d", test_cid)
	}
}

// TestGetProjectView tests the GetProjectView function of the projectview service.
// It verifies the behavior of the GetProjectView function with a non-existing projectview.
func TestGetProjectView_ProjectViewDoesNotExists(t *testing.T) {
	_, err := projectviewService.GetProjectView(test_pid+1, test_uid+1) // Should return error
	if err == nil {
		t.Error("No error returned in GetProjectView", err)
	}
}

// TestGetProjectViews tests the GetProjectViews function of the projectview service.
// It verifies the behavior of the GetProjectViews with existing projectviews.
func TestGetProjectViews_ProjectViewsExist(t *testing.T) {
	pv := projectviewService.GetProjectViews(test_uid)
	if len(pv) != test_total_models {
		t.Errorf("GetProjectViews should return an slice with %d elements", test_total_models)
	}
}

// TestGetProjectViews tests the GetProjectViews function of the projectview service.
// It verifies the behavior of the GetProjectViews with a non-existing projectviews.
func TestGetProjectViews_ProjectViewsDoNotExist(t *testing.T) {
	EmptyUid := 9999
	pv := projectviewService.GetProjectViews(EmptyUid)
	if len(pv) != 0 {
		t.Error("GetProjectViews should return an empty slice")
	}
}
