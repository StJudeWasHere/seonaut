package crawler

import (
	"errors"
	"net/url"
	"sync"

	"github.com/temoto/robotstxt"
)

type RobotsChecker struct {
	robotsMap map[string]*robotstxt.RobotsData
	rlock     *sync.RWMutex
	client    Client
}

func NewRobotsChecker(client Client) *RobotsChecker {
	return &RobotsChecker{
		robotsMap: make(map[string]*robotstxt.RobotsData),
		rlock:     &sync.RWMutex{},
		client:    client,
	}
}

// Returns true if the URL is blocked by robots.txt
func (r *RobotsChecker) IsBlocked(u *url.URL) bool {
	robot, err := r.getRobotsMap(u)
	if err != nil || robot == nil {
		return false
	}

	path := u.EscapedPath()
	if u.RawQuery != "" {
		path += "?" + u.Query().Encode()
	}

	return !robot.TestAgent(path, r.client.GetUA())
}

// Returns true if the robots.txt file exists and is valid
func (r *RobotsChecker) Exists(u *url.URL) bool {
	robot, err := r.getRobotsMap(u)
	if err != nil {
		return false
	}

	if robot == nil {
		return false
	}

	return true
}

// Returns a list of sitemaps found in the robots.txt file
func (r *RobotsChecker) GetSitemaps(u *url.URL) []string {
	robot, err := r.getRobotsMap(u)
	if err != nil || robot == nil {
		return []string{}
	}

	return robot.Sitemaps
}

// Returns a RobotsData checking if it has already been created and stored in the robotsMap
func (r *RobotsChecker) getRobotsMap(u *url.URL) (*robotstxt.RobotsData, error) {
	r.rlock.Lock()
	defer r.rlock.Unlock()

	robot, ok := r.robotsMap[u.Host]
	if ok {
		return robot, nil
	}

	resp, err := r.client.Get(u.Scheme + "://" + u.Host + "/robots.txt")
	if err != nil {
		r.robotsMap[u.Host] = nil
		return nil, err
	}
	defer resp.Response.Body.Close()

	if resp.Response.StatusCode != 200 {
		r.robotsMap[u.Host] = nil
		return nil, errors.New("robots.txt file does not exist")
	}

	robot, err = robotstxt.FromResponse(resp.Response)
	if err != nil {
		r.robotsMap[u.Host] = nil
		return nil, err
	}

	r.robotsMap[u.Host] = robot

	return robot, nil
}
