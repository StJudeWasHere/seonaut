package crawler

import (
	"net/http"
	"time"
)

type Client struct {
	userAgent string
	client    *http.Client
}

func NewClient(ua string) *Client {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return &Client{
		client:    httpClient,
		userAgent: ua,
	}
}

// Gets an URL and handles the response with the responseHandler method
func (c *Client) Get(u string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return &http.Response{}, err
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
