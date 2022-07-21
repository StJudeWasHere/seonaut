package user_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/user"
)

const (
	id       = 1
	email    = "user@example.com"
	password = "user_password"
)

var testUser = &user.User{
	Id:       id,
	Email:    email,
	Password: "$2a$10$REKj9zUr.reKqlKETKpq3OGGBhzyhQ2TQ3wOVnToroO.Qh3nuWziK",
}

type storage struct{}

func (s *storage) FindUserById(i int) *user.User {
	if i == testUser.Id {
		return testUser
	}

	return &user.User{}
}

func (s *storage) UserUpdatePassword(email, hashedPassword string) error {
	return nil
}

func (s *storage) UserSignup(e, p string) {}

func (s *storage) FindUserByEmail(e string) *user.User {
	if e == testUser.Email {
		return testUser
	}

	return &user.User{}
}

var service = user.NewService(&storage{})

func TestSignup(t *testing.T) {
	m := []struct {
		email    string
		password string
		err      bool
	}{
		{"new_user@example.com", "valid_password", false},
		{"new_user@example.com", "", true},
		{"", "valid_password", true},
		{"invalid_email", "valid_password", true},
		{email, password, true},
	}

	for _, v := range m {
		err := service.SignUp(v.email, v.password)
		if (v.err == true && err == nil) || (v.err == false && err != nil) {
			t.Errorf("Signup '%s' password '%s' should error %v", v.email, v.password, v.err)
		}
	}
}

func TestSigin(t *testing.T) {
	m := []struct {
		email    string
		password string
		err      bool
	}{
		{email, password, false},
		{"email", "invalid_password", true},
		{"not_user@example.com", password, true},
		{"", "", true},
	}

	for _, v := range m {
		_, err := service.SignIn(v.email, v.password)
		if (v.err == true && err == nil) || (v.err == false && err != nil) {
			t.Errorf("Signup '%s' password '%s' should error %v", v.email, v.password, v.err)
		}
	}
}

func TestFindById(t *testing.T) {
	u := service.FindById(id)
	if u != testUser {
		t.Error("User not found")
	}

	u = service.FindById(1009)
	if u == testUser {
		t.Error("User found")
	}
}
