package app

import (
	"database/sql"
)

const (
	MaxProjects         = 3
	AdvancedMaxProjects = 6
)

type User struct {
	Id              int
	Email           string
	Password        string
	Advanced        bool
	StripeSessionId sql.NullString
}

func (u *User) getMaxAllowedProjects() int {
	if u.Advanced {
		return AdvancedMaxProjects
	}

	return MaxProjects
}
