package container_test

import (
	"testing"

	"github.com/stjudewashere/seonaut/internal/container"
	"github.com/stjudewashere/seonaut/internal/models"
)

const (
	id       = 1
	email    = "user@example.com"
	password = "user_password"
)

// Fake user for testing purposes.
var testUser = &models.User{
	Id:       id,
	Email:    email,
	Password: "$2a$10$REKj9zUr.reKqlKETKpq3OGGBhzyhQ2TQ3wOVnToroO.Qh3nuWziK",
}

// Create a mock storage for the user service.
// The storage only contains the testUSer.
type userstorage struct{}

func (s *userstorage) FindUserById(i int) *models.User {
	if i == testUser.Id {
		return testUser
	}

	return &models.User{}
}

func (s *userstorage) UserUpdatePassword(email, hashedPassword string) error {
	return nil
}

func (s *userstorage) UserSignup(e, p string) (*models.User, error) {
	return &models.User{}, nil
}

func (s *userstorage) FindUserByEmail(e string) *models.User {
	if e == testUser.Email {
		return testUser
	}

	return &models.User{}
}
func (s *userstorage) DeleteUser(uid int) {}

var userService = container.NewUserService(&userstorage{})

// TestSignup tests the SignUp method of the user service.
// It verifies the behavior of the SignUp function for different input scenarios.
// For each test case, it calls the SignUp function with the provided email and password,
// and checks if the returned error matches the expected error value.
// If the expected error and the actual error do not match, it reports a test failure.
func TestSignup(t *testing.T) {
	m := []struct {
		email    string
		password string
		err      bool
	}{
		{"new_user@example.com", "valid_password", false}, // Valid email and password. Expects no error.
		{"new_user@example.com", "", true},                // Empty password. Expects an error.
		{"", "valid_password", true},                      // Empty email. Expects an error.
		{"invalid_email", "valid_password", true},         // Invalid email. Expects an error.
		{email, password, true},                           // Email that already exists. Expects an error.
	}

	for _, v := range m {
		_, err := userService.SignUp(v.email, v.password)
		if (v.err == true && err == nil) || (v.err == false && err != nil) {
			t.Errorf("Signup '%s' password '%s' should error %v", v.email, v.password, v.err)
		}
	}
}

// TestSigin tests the SignIn method of the user service.
// It verifies the behavior of the SignIn function for different input scenarios.
// For each test case, it calls the SignIn function with the provided email and password,
// and checks if the returned error matches the expected error value.
// If the expected error and the actual error do not match, it reports a test failure.
func TestSigin(t *testing.T) {
	m := []struct {
		email    string
		password string
		err      bool
	}{
		{email, password, false},                 // Valid email and password. Expects no error.
		{email, "invalid_password", true},        // Invalid password. Expects an error.
		{"not_user@example.com", password, true}, // Email does not exist. Expects an error.
		{"", "", true},                           // Empty email and password. Expects an error.
	}

	for _, v := range m {
		_, err := userService.SignIn(v.email, v.password)
		if (v.err == true && err == nil) || (v.err == false && err != nil) {
			t.Errorf("Signup '%s' password '%s' should error %v", v.email, v.password, v.err)
		}
	}
}

// TestFindById tests the FindById method of the user service.
// It verifies the behavior of the FindById function for different user ID scenarios.
func TestFindById(t *testing.T) {
	u := userService.FindById(id) // Should return the testUSer
	if u != testUser {
		t.Error("User not found")
	}

	unexistingUserId := 1009
	u = userService.FindById(unexistingUserId) // Should not return the testUSer
	if u == testUser {
		t.Error("User found")
	}
}
