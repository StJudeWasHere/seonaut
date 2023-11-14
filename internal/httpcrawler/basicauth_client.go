package httpcrawler

import (
	"net/http"
	"net/url"
	"time"
)

const (
	// HTTP client timeout in seconds.
	clientTimeOut = 10
)

type BasicAuthClient struct {
	Options *ClientOptions
	client  *http.Client
}

type ClientOptions struct {
	UserAgent        string
	BasicAuth        bool
	BasicAuthDomains []string
	AuthUser         string
	AuthPass         string
}

func NewClient(options *ClientOptions) *BasicAuthClient {
	httpClient := &http.Client{
		Timeout: clientTimeOut * time.Second,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &BasicAuthClient{
		client:  httpClient,
		Options: options,
	}
}

// Makes a request with the specified method.
func (c *BasicAuthClient) request(method, u string) (*http.Response, error) {
	req, err := http.NewRequest(method, u, nil)
	if err != nil {
		return &http.Response{}, err
	}

	domain, err := url.Parse(u)
	if err != nil {
		return &http.Response{}, err
	}

	if c.Options.BasicAuth && c.isBasicAuthDomain(domain.Host) {
		req.SetBasicAuth(c.Options.AuthUser, c.Options.AuthPass)
	}

	return c.Do(req)
}

// Returns true if the domain exists in the BasicAutDomains slice.
func (c *BasicAuthClient) isBasicAuthDomain(domain string) bool {
	for _, element := range c.Options.BasicAuthDomains {
		if element == domain {
			return true
		}
	}

	return false
}

// Makes a GET request to an URL and returns the http response or an error.
func (c *BasicAuthClient) Get(u string) (*http.Response, error) {
	return c.request(http.MethodGet, u)
}

// Makes a HEAD request to an URL and returns the http response or an error.
func (c *BasicAuthClient) Head(u string) (*http.Response, error) {
	return c.request(http.MethodHead, u)
}

// Does a request and returns its response and error.
// It sets the client's User-Agent as well as the BasicAuth details if they are available.
func (c *BasicAuthClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", c.Options.UserAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
