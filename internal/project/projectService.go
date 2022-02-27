package project

import (
	"net/url"
	"time"
)

type ProjectStore interface {
	FindProjectsByUser(int) []Project
	SaveProject(string, bool, bool, int)
	FindProjectById(id int, uid int) (Project, error)
}

type Project struct {
	Id              int
	URL             string
	Host            string
	IgnoreRobotsTxt bool
	UseJS           bool
	Created         time.Time
}

type ProjectService struct {
	store ProjectStore
}

func NewService(store ProjectStore) *ProjectService {
	return &ProjectService{
		store: store,
	}
}

func (s *ProjectService) GetProjects(userId int) []Project {
	return s.store.FindProjectsByUser(userId)
}

func (s *ProjectService) SaveProject(url string, ignoreRobotsTxt, useJavascript bool, userId int) {
	s.store.SaveProject(url, ignoreRobotsTxt, useJavascript, userId)
}

func (s *ProjectService) FindProject(id, uid int) (Project, error) {
	project, err := s.store.FindProjectById(id, uid)
	if err != nil {
		return project, nil
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		return project, err
	}

	project.Host = parsedURL.Host

	return project, nil
}
