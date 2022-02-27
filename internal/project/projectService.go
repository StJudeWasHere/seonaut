package project

import (
	"time"
)

type ProjectStore interface {
	FindProjectsByUser(int) []Project
	SaveProject(string, bool, bool, int)
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
