package project

import (
	"errors"
	"net/url"
	"time"
)

type Storage interface {
	SaveProject(*Project, int)
	DeleteProject(*Project)
	UpdateProject(p *Project) error
	FindProjectById(id int, uid int) (Project, error)
}

type Project struct {
	Id              int64
	URL             string
	Host            string
	IgnoreRobotsTxt bool
	FollowNofollow  bool
	IncludeNoindex  bool
	Created         time.Time
	CrawlSitemap    bool
	AllowSubdomains bool
	Deleting        bool
}

type ProjectService struct {
	storage Storage
}

func NewService(s Storage) *ProjectService {
	return &ProjectService{
		storage: s,
	}
}

// SaveProject stores a new project
func (s *ProjectService) SaveProject(project *Project, userId int) error {
	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		return err
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("Protocol not supported")
	}

	s.storage.SaveProject(project, userId)

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

// Delete a project
func (s *ProjectService) DeleteProject(p *Project) {
	s.storage.DeleteProject(p)
}

// Update project
func (s *ProjectService) UpdateProject(p *Project) error {
	return s.storage.UpdateProject(p)
}
