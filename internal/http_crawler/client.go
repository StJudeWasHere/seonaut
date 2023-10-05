package http_crawler

import (
	"net/http"
	"time"
)

type Client struct {
	Options *ClientOptions
	client  *http.Client
}

type ClientOptions struct {
	UserAgent string
	BasicAuth bool
	AuthUser  string
	AuthPass  string
}

func NewClient(options *ClientOptions) *Client {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Client{
		client:  httpClient,
		Options: options,
	}
}

// Makes a request with the specified method.
// It sets the client's User-Agent as well as the BasicAuth details if they are available.
func (c *Client) request(m, u string) (*http.Response, error) {
	req, err := http.NewRequest(m, u, nil)
	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Set("User-Agent", c.Options.UserAgent)
	if c.Options.BasicAuth {
		req.SetBasicAuth(c.Options.AuthUser, c.Options.AuthPass)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// Makes a GET request to an URL and returns the http response or an error.
func (c *Client) Get(u string) (*http.Response, error) {
	return c.request(http.MethodGet, u)
}

// Makes a HEAD request to an URL and returns the http response or an error.
func (c *Client) Head(u string) (*http.Response, error) {
	return c.request(http.MethodHead, u)
}
