package projectview_test

import (
	"errors"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/projectview"
)

const (
	test_uid = 1
	test_pid = 1
	test_cid = 1
	test_url = "https://example.org"
)

type Storage interface {
	FindProjectsByUser(int) []models.Project
	FindProjectById(id int, uid int) (models.Project, error)
	GetLastCrawl(*models.Project) models.Crawl
}

type testStorage struct{}

func (s *testStorage) FindProjectsByUser(uid int) []models.Project {
	return []models.Project{
		{},
		{},
	}
}

func (s *testStorage) FindProjectById(id int, uid int) (models.Project, error) {
	if id == test_pid && uid == test_uid {
		return models.Project{Id: test_pid, URL: test_url}, nil
	}

	return models.Project{}, errors.New("Test error")
}

func (s *testStorage) GetLastCrawl(*models.Project) models.Crawl {
	return models.Crawl{Id: test_cid}
}

var projectviewService = projectview.NewService(&testStorage{})

func TestGetProjectView(t *testing.T) {
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

	// Should return error
	_, err = projectviewService.GetProjectView(test_pid+1, test_uid+1)
	if err == nil {
		t.Error("No error returned in GetProjectView", err)
	}
}

func TestGetProjectViews(t *testing.T) {
	pv := projectviewService.GetProjectViews(1)
	if len(pv) != 2 {
		t.Error("len(pv) != 2")
	}
}
