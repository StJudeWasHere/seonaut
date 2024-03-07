package datastore

import (
	"log"

	"github.com/stjudewashere/seonaut/internal/models"
)

// UserSignup inserts a new user with the provided email and password into the database.
// It returns the inserted user and an error if the user could not be inserted.
func (ds *Datastore) UserSignup(user, password string) (*models.User, error) {
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

func (ds *Datastore) FindUserByEmail(email string) *models.User {
	u := models.User{}
	query := `
		SELECT
			id,
			email,
			password
		FROM users
		WHERE email = ? AND deleting = 0`

	row := ds.db.QueryRow(query, email)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}

func (ds *Datastore) FindUserById(id int) *models.User {
	u := models.User{}
	query := `
		SELECT
			id,
			email,
			password
		FROM users
		WHERE id = ? AND deleting = 0`

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

// DeleteUser deletes an existing user and all its associated projects and crawl data.
// It first updates the user to set the deleting field to 1. Once all the user data has
// been removed it proceeds to actually delete the user.
func (ds *Datastore) DeleteUser(uid int) {
	query := `UPDATE users SET deleting=1 WHERE id = ?`
	_, err := ds.db.Exec(query, uid)
	if err != nil {
		log.Printf("DeleteUser: update: pid %d %v\n", uid, err)
		return
	}

	projects := ds.FindProjectsByUser(uid)
	for _, p := range projects {
		ds.DeleteProject(&p)
	}

	deleteQuery := "DELETE FROM users WHERE id = ?"
	_, err = ds.db.Exec(deleteQuery, uid)
	if err != nil {
		log.Printf("DeleteUser: %v", err)
	}
}
