package crawler_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/stjudewashere/seonaut/internal/crawler"
)

type mockClient struct {
	lastRequest *http.Request
	ForceError  bool
}

func (m *mockClient) Do(req *http.Request) (*http.Response, error) {
	m.lastRequest = req

	if m.ForceError {
		return nil, fmt.Errorf("mock error")
	}

	return &http.Response{
		StatusCode: http.StatusOK,
	}, nil
}

// Test user agent in Get requests.
func TestGetUserAgent(t *testing.T) {
	testUA := "TEST_UA"
	options := &crawler.ClientOptions{
		UserAgent: testUA,
	}

	mockClient := &mockClient{}
	client := crawler.NewBasicClient(options, mockClient)

	_, err := client.Get("http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockClient.lastRequest == nil {
		t.Fatalf("expected a request to be made, but it was nil")
	}

	ua := mockClient.lastRequest.Header.Get("User-Agent")
	if ua != testUA {
		t.Errorf("expected User-Agent to be %s, got '%s'", testUA, ua)
	}
}

// Test user agent in Head requests.
func TestHeadUserAgent(t *testing.T) {
	testUA := "TEST_UA"
	options := &crawler.ClientOptions{
		UserAgent: testUA,
	}

	mockClient := &mockClient{}
	client := crawler.NewBasicClient(options, mockClient)

	_, err := client.Head("http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockClient.lastRequest == nil {
		t.Fatalf("expected a request to be made, but it was nil")
	}

	ua := mockClient.lastRequest.Header.Get("User-Agent")
	if ua != testUA {
		t.Errorf("expected User-Agent to be %s, got '%s'", testUA, ua)
	}
}

// Test Basic Auth headers are sent for valid domains.
func TestBasicAuthHeadersSent(t *testing.T) {
	authUser := "user"
	authPass := "pass"
	options := &crawler.ClientOptions{
		UserAgent:        "TEST_UA",
		BasicAuthDomains: []string{"example.com"},
		AuthUser:         authUser,
		AuthPass:         authPass,
	}

	mockClient := &mockClient{}
	client := crawler.NewBasicClient(options, mockClient)

	_, err := client.Get("http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockClient.lastRequest == nil {
		t.Fatalf("expected a request to be made, but it was nil")
	}

	authHeader := mockClient.lastRequest.Header.Get("Authorization")
	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(authUser+":"+authPass))
	if authHeader != expectedAuth {
		t.Errorf("expected Authorization header to be %s, got '%s'", expectedAuth, authHeader)
	}
}

// Test Basic Auth headers are not sent for domains not in the BasicAuthDomains list.
func TestBasicAuthHeadersNotSent(t *testing.T) {
	options := &crawler.ClientOptions{
		UserAgent:        "TEST_UA",
		BasicAuthDomains: []string{}, // No domains for Basic Auth
		AuthUser:         "user",
		AuthPass:         "pass",
	}

	mockClient := &mockClient{}
	client := crawler.NewBasicClient(options, mockClient)

	_, err := client.Get("http://example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mockClient.lastRequest == nil {
		t.Fatalf("expected a request to be made, but it was nil")
	}

	authHeader := mockClient.lastRequest.Header.Get("Authorization")
	if authHeader != "" {
		t.Errorf("expected no Authorization header, got '%s'", authHeader)
	}
}

// Test client error.
func TestHTTPError(t *testing.T) {
	options := &crawler.ClientOptions{}

	mockClient := &mockClient{ForceError: true}
	client := crawler.NewBasicClient(options, mockClient)

	_, err := client.Get("http://example.com")
	if err == nil {
		t.Fatal("expected an error, got none")
	}
}
