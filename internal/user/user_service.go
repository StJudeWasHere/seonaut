package user

import (
	"context"
	"errors"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	FindUserById(int) *User
	UserSignup(string, string) (*User, error)
	FindUserByEmail(string) *User
	UserUpdatePassword(email, hashedPassword string) error
}

type Service struct {
	store UserStore
}

type User struct {
	Id       int
	Email    string
	Password string
}

func NewService(s UserStore) *Service {
	return &Service{
		store: s,
	}
}

// FindById returns a by its Id.
func (s *Service) FindById(id int) *User {
	return s.store.FindUserById(id)
}

// SignUp validates the user email and password, if they are both valid creates a password hash
// before storing it. If the storage is succesful it returns the new user.
func (s *Service) SignUp(email, password string) (*User, error) {
	u := s.store.FindUserByEmail(email)
	if u.Id != 0 {
		return nil, errors.New("user already exists")
	}

	if len(password) < 1 {
		return nil, errors.New("invalid password")
	}

	_, err := mail.ParseAddress(email)
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
func (s *Service) SignIn(email, password string) (*User, error) {
	u := s.store.FindUserByEmail(email)
	if u.Id == 0 {
		return nil, errors.New("user does not exist")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("incorrect password")
	}

	return u, nil
}

// UpdatePassword updates the password for the user with the given email.
// It validates the new password and generates a hashed password using bcrypt before storing it.
func (s *Service) UpdatePassword(email, password string) error {
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

// GetUserFromContext takes a context as input and retrieves the associated User value from it, if present.
func (s *Service) GetUserFromContext(c context.Context) (*User, bool) {
	v := c.Value("user")
	user, ok := v.(*User)

	return user, ok
}

// SetUserToContext takes a User and a context as input and returns a new context with the given
// user value set.
func (s *Service) SetUserToContext(user *User, c context.Context) context.Context {
	return context.WithValue(c, "user", user)
}
