package http_crawler

import (
	"net/http"
	"time"
)

type Client struct {
	options *ClientOptions
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
		options: options,
	}
}

// Makes a GET request to an URL and returns the http response or an error.
// It sets the client's User-Agent as well as the BasicAuth details if they are available.
func (c *Client) Get(u string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Set("User-Agent", c.options.UserAgent)
	if c.options.BasicAuth {
		req.SetBasicAuth(c.options.AuthUser, c.options.AuthPass)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
