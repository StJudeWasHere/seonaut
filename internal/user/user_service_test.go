package user_test

import (
	"context"
	"testing"

	"github.com/stjudewashere/seonaut/internal/user"
)

const (
	id       = 1
	email    = "user@example.com"
	password = "user_password"
)

// Fake user for testing purposes.
var testUser = &user.User{
	Id:       id,
	Email:    email,
	Password: "$2a$10$REKj9zUr.reKqlKETKpq3OGGBhzyhQ2TQ3wOVnToroO.Qh3nuWziK",
}

// Create a mock storage for the user service.
// The storage only contains the testUSer.
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

func (s *storage) UserSignup(e, p string) (*user.User, error) {
	return &user.User{}, nil
}

func (s *storage) FindUserByEmail(e string) *user.User {
	if e == testUser.Email {
		return testUser
	}

	return &user.User{}
}

var service = user.NewService(&storage{})

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
		_, err := service.SignUp(v.email, v.password)
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
		_, err := service.SignIn(v.email, v.password)
		if (v.err == true && err == nil) || (v.err == false && err != nil) {
			t.Errorf("Signup '%s' password '%s' should error %v", v.email, v.password, v.err)
		}
	}
}

// TestFindById tests the FindById method of the user service.
// It verifies the behavior of the FindById function for different user ID scenarios.
func TestFindById(t *testing.T) {
	u := service.FindById(id) // Should return the testUSer
	if u != testUser {
		t.Error("User not found")
	}

	unexistingUserId := 1009
	u = service.FindById(unexistingUserId) // Should not return the testUSer
	if u == testUser {
		t.Error("User found")
	}
}

// TestSetUserToContext verifies the behavior of the SetUserToContext function in the user service.
// It tests whether the function correctly sets the user value in the context.
func TestSetUserToContext(t *testing.T) {
	ctx := context.Background()
	newCtx := service.SetUserToContext(testUser, ctx)

	resultUser, ok := newCtx.Value("user").(*user.User)

	if !ok || resultUser != testUser {
		t.Errorf("SetUserToContext did not set the user correctly, got %v, expected %v", resultUser, testUser)
	}
}

// TestGetUserFromContext_UserExists verifies the behavior of the GetUserFromContext function
// in the user service when a user exists in the context.
func TestGetUserFromContext_UserExists(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user", testUser)

	resultUser, ok := service.GetUserFromContext(ctx)

	if !ok || resultUser != testUser {
		t.Errorf("GetUserFromContext returned user %v, expected user %v", resultUser, testUser)
	}
}

// TestGetUserFromContext_UserExists verifies the behavior of the GetUserFromContext function
// in the user service when a user does not exists in the context.
func TestGetUserFromContext_UserDoesNotExist(t *testing.T) {
	ctx := context.Background()
	resultUser, ok := service.GetUserFromContext(ctx)

	if ok || resultUser != nil {
		t.Error("GetUserFromContext returned a user when it shouldn't have")
	}
}
