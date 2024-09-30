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

// Create a mock repository for the user service.
// The repository only contains the testUSer.
type userTestRepository struct {
	userDeleted bool
}

var ErrUnexistingUser = errors.New("user does not exist")

func (s *userTestRepository) UserUpdatePassword(email, hashedPassword string) error {
	if email != testUser.Email {
		return ErrUnexistingUser
	}

	return nil
}
func (s *userTestRepository) UserSignup(e, p string) (*models.User, error) {
	return &models.User{}, nil
}
func (s *userTestRepository) FindUserByEmail(e string) (*models.User, error) {
	if e == testUser.Email {
		return testUser, nil
	}

	return &models.User{}, services.ErrUnexistingUser
}
func (s *userTestRepository) DeleteUser(u *models.User) error {
	if u == testUser {
		s.userDeleted = true
	}

	return nil
}
func (s *userTestRepository) DisableUser(*models.User) error {
	return nil
}

var userService = services.NewUserService(&userTestRepository{})

// TestSignup tests the SignUp method of the user service.
// It verifies the behavior of the SignUp function for different input scenarios.
// For each test case, it calls the SignUp function with the provided email and password,
// and checks if the returned error matches the expected error value.
// If the expected error and the actual error do not match, it reports a test failure.
func TestSignup(t *testing.T) {
	m := []struct {
		email    string
		password string
		err      error
	}{
		{"new_user@example.com", "valid_password", nil},               // Valid email and password. Expects no error.
		{"new_user@example.com", "", services.ErrInvalidPassword},     // Empty password. Expects an error.
		{"", "valid_password", services.ErrInvalidEmail},              // Empty email. Expects an error.
		{"invalid_email", "valid_password", services.ErrInvalidEmail}, // Invalid email. Expects an error.
		{email, password, services.ErrUserExists},                     // Email that already exists. Expects an error.
	}

	for _, v := range m {
		_, err := userService.SignUp(v.email, v.password)
		if err != v.err {
			t.Errorf("Signup '%s' password '%s' should error '%v': '%v'", v.email, v.password, v.err, err)
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
		email           string
		currentPassword string
		newPassword     string
		err             error
	}{
		{testUser.Email, password, "valid_password", nil},                       // Valid email and password. Expects no error.
		{testUser.Email, password, "", services.ErrInvalidPassword},             // Empty password. Expects an error.
		{testUser.Email, "", "valid_password", services.ErrIncorrectPassword},   // Current password is invalid.
		{"not_user@example.com", password, "valid_password", ErrUnexistingUser}, // User does not exist. Expects an error.
	}

	for _, v := range m {
		u := &models.User{
			Email:    v.email,
			Password: testUser.Password,
		}
		err := userService.UpdatePassword(u, v.currentPassword, v.newPassword)
		if err != v.err {
			t.Errorf("Signup '%s' password '%s' should error '%v': '%v'", v.email, v.newPassword, v.err, err)
		}
	}
}
