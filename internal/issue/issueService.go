package issue

const (
	Critical = iota + 1
	Alert
	Warning
)

type IssueStore interface {
	FindIssues(int) map[string]IssueGroup
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
	Groups   map[string]IssueGroup
	Critical int
	Alert    int
	Warning  int
}

func NewService(s IssueStore) *IssueService {
	return &IssueService{
		store: s,
	}
}

func (s *IssueService) GetIssuesCount(crawlId int) *IssueCount {
	c := &IssueCount{}

	c.Groups = s.store.FindIssues(crawlId)

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
