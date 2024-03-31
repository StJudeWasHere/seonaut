package repository

import (
	"database/sql"

	"github.com/stjudewashere/seonaut/internal/models"
)

type UserRepository struct {
	DB *sql.DB
}

// UserSignup inserts a new user with the provided email and password into the database.
// It returns the inserted user and an error if the user could not be inserted.
func (ds *UserRepository) UserSignup(email, password string) (*models.User, error) {
	query := `INSERT INTO users (email, password, created) VALUES (?, ?, NOW())`
	stmt, _ := ds.DB.Prepare(query)
	defer stmt.Close()

	_, err := stmt.Exec(email, password)
	if err != nil {
		return nil, err
	}

	// Retrieve and return the inserted user
	return ds.FindUserByEmail(email)
}

// FindUserByEmail returns the user that matches the email address. It also returns
// an error if the user does not exist. The user must not be in 'deleting' state.
func (ds *UserRepository) FindUserByEmail(email string) (*models.User, error) {
	u := &models.User{}
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
		return u, err
	}

	return u, nil
}

// UserUpdatePassword sets a new password for the user matching the email address.
func (ds *UserRepository) UserUpdatePassword(email, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = ?
		WHERE email = ?
	`

	_, err := ds.DB.Exec(query, hashedPassword, email)

	return err
}

// DisableUser sets the user in 'deleting' state, which makes it inactive.
func (ds *UserRepository) DisableUser(user *models.User) error {
	query := `UPDATE users SET deleting=1 WHERE id = ?`
	_, err := ds.DB.Exec(query, user.Id)

	return err
}

// DeleteUser removes the user from the database.
func (ds *UserRepository) DeleteUser(user *models.User) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := ds.DB.Exec(query, user.Id)

	return err
}
