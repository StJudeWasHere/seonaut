package user

import (
	"context"
	"errors"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	FindUserById(int) *User
	UserSignup(string, string)
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

// SignUp validates the user email and password and saves a new
// user in the user storage.
func (s *Service) SignUp(email, password string) error {
	u := s.store.FindUserByEmail(email)
	if u.Id != 0 {
		return errors.New("user already exists")
	}

	if len(password) < 1 {
		return errors.New("invalid password")
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.New("invalid email")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	s.store.UserSignup(email, string(hashedPassword))

	return nil
}

// SignIn checks if user credencials are correct to sign in a user.
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

// Sets a new password for the user identified with the email address.
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

// Gets a User from the given Context
func (s *Service) GetUserFromContext(c context.Context) (*User, bool) {
	v := c.Value("user")
	user, ok := v.(*User)
	return user, ok
}

// Returns a Context with the given User
func (s *Service) SetUserToContext(user *User, c context.Context) context.Context {
	return context.WithValue(c, "user", user)
}
