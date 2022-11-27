package projectview_test

import (
	"errors"
	"testing"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/project"
	"github.com/stjudewashere/seonaut/internal/projectview"
)

const (
	test_uid = 1
	test_pid = 1
	test_cid = 1
	test_url = "https://example.org"
)

type Storage interface {
	FindProjectsByUser(int) []project.Project
	FindProjectById(id int, uid int) (project.Project, error)
	GetLastCrawl(*project.Project) crawler.Crawl
}

type testStorage struct{}

func (s *testStorage) FindProjectsByUser(uid int) []project.Project {
	return []project.Project{
		{},
		{},
	}
}

func (s *testStorage) FindProjectById(id int, uid int) (project.Project, error) {
	if id == test_pid && uid == test_uid {
		return project.Project{Id: test_pid, URL: test_url}, nil
	}

	return project.Project{}, errors.New("Test error")
}

func (s *testStorage) GetLastCrawl(*project.Project) crawler.Crawl {
	return crawler.Crawl{Id: test_cid}
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
