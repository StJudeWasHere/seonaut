package project

import (
	"errors"
	"net/url"
	"strings"
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
	FollowNofollow  bool
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

func (s *ProjectService) SaveProject(u string, ignoreRobotsTxt, followNofollow bool, userId int) error {
	u = strings.TrimSpace(u)
	p, err := url.ParseRequestURI(u)
	if err != nil {
		return err
	}

	if p.Scheme != "http" && p.Scheme != "https" {
		return errors.New("Protocol not supported")
	}

	s.store.SaveProject(u, ignoreRobotsTxt, followNofollow, userId)

	return nil
}

func (s *ProjectService) FindProject(id, uid int) (Project, error) {
	project, err := s.store.FindProjectById(id, uid)
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
