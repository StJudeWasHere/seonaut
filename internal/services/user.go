package services

import (
	"errors"
	"net/mail"

	"github.com/stjudewashere/seonaut/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	// Error returned when the email is not a valid email.
	ErrInvalidEmail = errors.New("user service: invalid email")

	// Error returned when the password does not follow the password criteria.
	ErrInvalidPassword = errors.New("user service: invalid password")

	// Error returned when the user we are authenticating does not exist.
	ErrUnexistingUser = errors.New("user service: user does not exist")

	// Error returned when the password is incorrect for the user we are authenticating.
	ErrIncorrectPassword = errors.New("user service: incorrect password")

	// Error returned when trying to create a user that is already signed up.
	ErrUserExists = errors.New("user service: user already exists")
)

type (
	UserServiceStorage interface {
		UserSignup(email, hashedPassword string) (*models.User, error)
		FindUserByEmail(email string) (*models.User, error)
		UserUpdatePassword(email, hashedPassword string) error
		DeleteUser(user *models.User) error
		DisableUser(user *models.User) error

		DeleteProjectCrawls(project *models.Project)

		DeleteProject(project *models.Project)
		FindProjectsByUser(userId int) []models.Project
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
		return nil, ErrUserExists
	}

	if !s.validPassword(password) {
		return nil, ErrInvalidPassword
	}

	_, err = mail.ParseAddress(email)
	if err != nil {
		return nil, ErrInvalidEmail
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
		return nil, ErrUnexistingUser
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, ErrIncorrectPassword
	}

	return u, nil
}

// UpdatePassword updates the password for the user with the given email.
// It validates the new password and generates a hashed password using bcrypt before storing it.
func (s *UserService) UpdatePassword(user *models.User, currentPassword, newPassword string) error {
	if !s.validPassword(newPassword) {
		return ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword)); err != nil {
		return ErrIncorrectPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	err = s.store.UserUpdatePassword(user.Email, string(hashedPassword))
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

// Validate the password to make sure it follows certain criteria.
func (s *UserService) validPassword(password string) bool {
	return len(password) > 1
}
