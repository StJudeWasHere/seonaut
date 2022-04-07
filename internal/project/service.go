package project

import (
	"errors"
	"net/url"
	"strings"
	"time"
)

type Storage interface {
	SaveProject(string, bool, bool, bool, int)
	FindProjectById(id int, uid int) (Project, error)
}

type Project struct {
	Id              int
	URL             string
	Host            string
	IgnoreRobotsTxt bool
	FollowNofollow  bool
	IncludeNoindex  bool
	Created         time.Time
}

type ProjectService struct {
	storage Storage
}

func NewService(s Storage) *ProjectService {
	return &ProjectService{
		storage: s,
	}
}

// SaveProject stores a new project for a specific user with all the specified options.
func (s *ProjectService) SaveProject(u string, ignoreRobotsTxt, followNofollow, includeNoindex bool, userId int) error {
	p, err := url.ParseRequestURI(strings.TrimSpace(u))
	if err != nil {
		return err
	}

	if p.Scheme != "http" && p.Scheme != "https" {
		return errors.New("Protocol not supported")
	}

	s.storage.SaveProject(p.String(), ignoreRobotsTxt, followNofollow, includeNoindex, userId)

	return nil
}

// Return a project specified by id and user.
// It populates the Host field from the project's URL.
func (s *ProjectService) FindProject(id, uid int) (Project, error) {
	project, err := s.storage.FindProjectById(id, uid)
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
