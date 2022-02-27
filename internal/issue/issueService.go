package issue

const (
	Critical = iota + 1
	Alert
	Warning
)

type IssueStore interface {
	FindIssues(int) map[string]IssueGroup
	CountByMediaType(int) CountList
	CountByStatusCode(int) CountList
}

type IssueService struct {
	store IssueStore
}

type IssueGroup struct {
	ErrorType string
	Priority  int
	Count     int
}

type IssueCount struct {
	Groups      map[string]IssueGroup
	Critical    int
	Alert       int
	Warning     int
	MediaCount  CountList
	StatusCount CountList
}

func NewService(s IssueStore) *IssueService {
	return &IssueService{
		store: s,
	}
}

func (s *IssueService) GetIssuesCount(crawlID int) *IssueCount {
	c := &IssueCount{
		Groups:      s.store.FindIssues(crawlID),
		MediaCount:  s.store.CountByMediaType(crawlID),
		StatusCount: s.store.CountByStatusCode(crawlID),
	}

	for _, v := range c.Groups {
		switch v.Priority {
		case Critical:
			c.Critical += v.Count
		case Alert:
			c.Alert += v.Count
		case Warning:
			c.Warning += v.Count
		}
	}

	return c
}
