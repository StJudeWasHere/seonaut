package datastore

import (
	"log"

	"github.com/stjudewashere/seonaut/internal/user"
)

// UserSignup inserts a new user with the provided email and password into the database.
// It returns the inserted user and an error if the user could not be inserted.
func (ds *Datastore) UserSignup(user, password string) (*user.User, error) {
	query := `INSERT INTO users (email, password, created) VALUES (?, ?, NOW())`
	stmt, _ := ds.db.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(user, password)
	if err != nil {
		return nil, err
	}

	// Retrieve the inserted user
	u := ds.FindUserByEmail(user)

	return u, nil
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

func (ds *Datastore) UserUpdatePassword(email, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = ?
		WHERE email = ?
	`

	_, err := ds.db.Exec(query, hashedPassword, email)

	return err
}
