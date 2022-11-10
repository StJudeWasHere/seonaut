package crawler

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

// Gets an URL and handles the response with the responseHandler method
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
