package services_test

import (
	"errors"
	"testing"

	"github.com/stjudewashere/seonaut/internal/models"
	"github.com/stjudewashere/seonaut/internal/services"
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

func (s *userstorage) UserUpdatePassword(email, hashedPassword string) error {
	if email != testUser.Email {
		return errors.New("user does not exist")
	}

	return nil
}
func (s *userstorage) UserSignup(e, p string) (*models.User, error) {
	return &models.User{}, nil
}
func (s *userstorage) FindUserByEmail(e string) (*models.User, error) {
	if e == testUser.Email {
		return testUser, nil
	}

	return &models.User{}, errors.New("user does not exist")
}
func (s *userstorage) DeleteUser(*models.User) error {
	return nil
}
func (s *userstorage) DisableUser(*models.User) error {
	return nil
}
func (p *userstorage) DeleteProject(*models.Project) {}
func (p *userstorage) FindProjectsByUser(uid int) []models.Project {
	return []models.Project{}
}
func (p *userstorage) DeleteProjectCrawls(*models.Project) {}

var userService = services.NewUserService(&userstorage{})

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

// TestUpdatePassword tests the UpdatePassword method of the user service.
// It verifies the behavior of the UpdatePassword function for different input scenarios.
// For each test case, it calls the UpdatePassword function with the provided email and password,
// and checks if the returned error matches the expected error value.
// If the expected error and the actual error do not match, it reports a test failure.
func TestUpdatePassword(t *testing.T) {
	m := []struct {
		email    string
		password string
		err      bool
	}{
		{testUser.Email, "valid_password", false},        // Valid email and password. Expects no error.
		{testUser.Email, "", true},                       // Empty password. Expects an error.
		{"not_user@example.com", "valid_password", true}, // User does not exist. Expects an error.
	}

	for _, v := range m {
		err := userService.UpdatePassword(v.email, v.password)
		if (v.err == true && err == nil) || (v.err == false && err != nil) {
			t.Errorf("Signup '%s' password '%s' should error %v", v.email, v.password, v.err)
		}
	}
}
