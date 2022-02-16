package main

import (
	"database/sql"
)

type User struct {
	Id              int
	Email           string
	Password        string
	Advanced        bool
	StripeSessionId sql.NullString
}

func (u User) IsAdvanced() bool {
	return u.Advanced
}
