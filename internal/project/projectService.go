package project

import (
	"net/url"
	"time"
)

type ProjectStore interface {
	FindProjectsByUser(int) []Project
	SaveProject(string, bool, int)
	FindProjectById(id int, uid int) (Project, error)

	GetLastCrawl(*Project) Crawl
}

type Project struct {
	Id              int
	URL             string
	Host            string
	IgnoreRobotsTxt bool
	Created         time.Time
}

type ProjectService struct {
	store ProjectStore
}

type ProjectView struct {
	Project Project
	Crawl   Crawl
}

func NewService(store ProjectStore) *ProjectService {
	return &ProjectService{
		store: store,
	}
}

func (s *ProjectService) GetProjects(userId int) []Project {
	return s.store.FindProjectsByUser(userId)
}

func (s *ProjectService) SaveProject(url string, ignoreRobotsTxt bool, userId int) {
	s.store.SaveProject(url, ignoreRobotsTxt, userId)
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

func (s *ProjectService) GetProjectView(id, uid int) (*ProjectView, error) {
	v := &ProjectView{}

	p, err := s.FindProject(id, uid)
	if err != nil {
		return v, err
	}

	c := s.store.GetLastCrawl(&p)

	v.Project = p
	v.Crawl = c

	return v, nil
}

func (s *ProjectService) GetProjectViews(uid int) []ProjectView {
	projects := s.GetProjects(uid)

	var views []ProjectView
	for _, p := range projects {
		pv := ProjectView{
			Project: p,
			Crawl:   s.store.GetLastCrawl(&p),
		}
		views = append(views, pv)
	}

	return views
}
