package models

type (
	IssuesGroupView struct {
		ProjectView *ProjectView
		IssueCount  *IssueCount
	}

	IssuesView struct {
		ProjectView   *ProjectView
		Eid           string
		PaginatorView PaginatorView
	}
)
