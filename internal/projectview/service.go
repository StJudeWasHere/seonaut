package projectview

import (
	"net/url"

	"github.com/stjudewashere/seonaut/internal/crawler"
	"github.com/stjudewashere/seonaut/internal/project"
)

type Storage interface {
	FindProjectsByUser(int) []project.Project
	FindProjectById(id int, uid int) (project.Project, error)

	GetLastCrawl(*project.Project) crawler.Crawl
}

type Service struct {
	storage Storage
}

type ProjectView struct {
	Project project.Project
	Crawl   crawler.Crawl
}

func NewService(s Storage) *Service {
	return &Service{
		storage: s,
	}
}

func (s *Service) GetProjectView(id, uid int) (*ProjectView, error) {
	v := &ProjectView{}

	project, err := s.storage.FindProjectById(id, uid)
	if err != nil {
		return nil, err
	}

	parsedURL, err := url.Parse(project.URL)
	if err != nil {
		return nil, err
	}

	project.Host = parsedURL.Host

	c := s.storage.GetLastCrawl(&project)

	v.Project = project
	v.Crawl = c

	return v, nil
}

func (s *Service) GetProjectViews(uid int) []ProjectView {
	projects := s.storage.FindProjectsByUser(uid)

	var views []ProjectView
	for _, p := range projects {
		pv := ProjectView{
			Project: p,
			Crawl:   s.storage.GetLastCrawl(&p),
		}
		views = append(views, pv)
	}

	return views
}
