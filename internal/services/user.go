package services

import (
	"errors"
	"net/mail"

	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type (
	UserServiceStorage interface {
		UserSignup(string, string) (*models.User, error)
		FindUserByEmail(string) (*models.User, error)
		UserUpdatePassword(email, hashedPassword string) error
		DeleteUser(*models.User) error
		DisableUser(*models.User) error

		DeleteProjectCrawls(*models.Project)

		DeleteProject(*models.Project)
		FindProjectsByUser(uid int) []models.Project
	}

	UserService struct {
		store UserServiceStorage
	}
)

func NewUserService(s UserServiceStorage) *UserService {
	return &UserService{
		store: s,
	}
}

// SignUp validates the user email and password, if they are both valid creates a password hash
// before storing it. If the storage is succesful it returns the new user.
func (s *UserService) SignUp(email, password string) (*models.User, error) {
	_, err := s.store.FindUserByEmail(email)
	if err == nil {
		return nil, errors.New("user already exists")
	}

	if len(password) < 1 {
		return nil, errors.New("invalid password")
	}

	_, err = mail.ParseAddress(email)
	if err != nil {
		return nil, errors.New("invalid email")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return s.store.UserSignup(email, string(hashedPassword))
}

// SignIn validates the provided email and password combination for user authentication.
// It compares the provided password with the user's hashed password.
// If the passwords do not match, it returns an error.
func (s *UserService) SignIn(email, password string) (*models.User, error) {
	u, err := s.store.FindUserByEmail(email)
	if err != nil {
		return nil, errors.New("user does not exist")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("incorrect password")
	}

	return u, nil
}

// UpdatePassword updates the password for the user with the given email.
// It validates the new password and generates a hashed password using bcrypt before storing it.
func (s *UserService) UpdatePassword(email, password string) error {
	if len(password) < 1 {
		return errors.New("invalid password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.store.UserUpdatePassword(email, string(hashedPassword))
	if err != nil {
		return err
	}

	return nil
}

// Delete a User and all its associated projects and crawl data.
func (s *UserService) DeleteUser(user *models.User) {
	s.store.DisableUser(user)
	go func() {
		projects := s.store.FindProjectsByUser(user.Id)
		for _, p := range projects {
			s.store.DeleteProjectCrawls(&p)
			s.store.DeleteProject(&p)
		}

		s.store.DeleteUser(user)
	}()
}
