package user

import (
	"errors"
	"net/mail"

	"golang.org/x/crypto/bcrypt"
)

type UserStore interface {
	FindUserById(int) *User
	UserSignup(string, string)
	FindUserByEmail(string) *User
}

type UserService struct {
	store UserStore
}

type User struct {
	Id       int
	Email    string
	Password string
}

func NewService(s UserStore) *UserService {
	return &UserService{
		store: s,
	}
}

// FindById returns a by its Id.
func (s *UserService) FindById(id int) *User {
	return s.store.FindUserById(id)
}

// SignUp validates the user email and password and saves a new
// user in the user storage.
func (s *UserService) SignUp(email, password string) error {
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
func (s *UserService) SignIn(email, password string) (*User, error) {
	u := s.store.FindUserByEmail(email)
	if u.Id == 0 {
		return nil, errors.New("user does not exist")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, errors.New("incorrect password")
	}

	return u, nil
}
