package datastore

import (
	"log"

	"github.com/stjudewashere/seonaut/internal/user"
)

func (ds *Datastore) UserSignup(user, password string) {
	query := `INSERT INTO users (email, password, created) VALUES (?, ?, NOW())`
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
