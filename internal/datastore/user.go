package datastore

import (
	"database/sql"
	"log"

	"github.com/mnlg/lenkrr/internal/user"
)

func (ds *Datastore) EmailExists(email string) bool {
	query := `SELECT exists (SELECT id FROM users WHERE email = ?)`
	var exists bool
	err := ds.db.QueryRow(query, email).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Error checking if email exists '%s' %v", email, err)
	}

	return exists
}

func (ds *Datastore) UserSignup(user, password string) {
	query := `INSERT INTO users (email, password) VALUES (?, ?)`
	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(user, password)
	if err != nil {
		log.Printf("UserSignup: %v\n", err)
	}
}

func (ds *Datastore) FindUserByEmail(email string) *user.User {
	u := user.User{}
	query := `
		SELECT
			id,
			email,
			password
		FROM users
		WHERE email = ?`

	row := ds.db.QueryRow(query, email)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}

func (ds *Datastore) FindUserById(id int) *user.User {
	u := user.User{}
	query := `
		SELECT
			id,
			email,
			password
		FROM users
		WHERE id = ?`

	row := ds.db.QueryRow(query, id)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}
