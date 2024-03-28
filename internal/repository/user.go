package repository

import (
	"database/sql"
	"log"

	"github.com/stjudewashere/seonaut/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

// UserSignup inserts a new user with the provided email and password into the database.
// It returns the inserted user and an error if the user could not be inserted.
func (ds *UserRepository) UserSignup(user, password string) (*models.User, error) {
	query := `INSERT INTO users (email, password, created) VALUES (?, ?, NOW())`
	stmt, _ := ds.DB.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(user, password)
	if err != nil {
		return nil, err
	}

	// Retrieve the inserted user
	u := ds.FindUserByEmail(user)

	return u, nil
}

func (ds *UserRepository) FindUserByEmail(email string) *models.User {
	u := models.User{}
	query := `
		SELECT
			id,
			email,
			password
		FROM users
		WHERE email = ? AND deleting = 0`

	row := ds.DB.QueryRow(query, email)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}

func (ds *UserRepository) FindUserById(id int) *models.User {
	u := models.User{}
	query := `
		SELECT
			id,
			email,
			password
		FROM users
		WHERE id = ? AND deleting = 0`

	row := ds.DB.QueryRow(query, id)
	err := row.Scan(&u.Id, &u.Email, &u.Password)
	if err != nil {
		log.Println(err)
		return &u
	}

	return &u
}

func (ds *UserRepository) UserUpdatePassword(email, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = ?
		WHERE email = ?
	`

	_, err := ds.DB.Exec(query, hashedPassword, email)

	return err
}

func (ds *UserRepository) DisableUser(uid int) {
	query := `UPDATE users SET deleting=1 WHERE id = ?`
	_, err := ds.DB.Exec(query, uid)
	if err != nil {
		log.Printf("DeleteUser: update: pid %d %v\n", uid, err)
	}
}

func (ds *UserRepository) DeleteUser(uid int) {
	query := "DELETE FROM users WHERE id = ?"
	_, err := ds.DB.Exec(query, uid)
	if err != nil {
		log.Printf("DeleteUser: %v", err)
	}
}
