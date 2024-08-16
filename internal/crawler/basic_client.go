package crawler

import (
	"math"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"time"
)

type HTTPRequester interface {
	Do(req *http.Request) (*http.Response, error)
}

type BasicClient struct {
	Options *ClientOptions
	client  HTTPRequester
}

type ClientOptions struct {
	UserAgent        string
	BasicAuthDomains []string
	AuthUser         string
	AuthPass         string
}

func NewBasicClient(options *ClientOptions, client HTTPRequester) *BasicClient {
	return &BasicClient{
		Options: options,
		client:  client,
	}
}

// Makes a request with the method specified in the method parameter to the specified URL.
func (c *BasicClient) request(method, urlStr string) (*ClientResponse, error) {
	req, err := http.NewRequest(method, urlStr, nil)
	if err != nil {
		return nil, err
	}

	domain, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	if c.Options.AuthUser != "" && c.isBasicAuthDomain(domain.Host) {
		req.SetBasicAuth(c.Options.AuthUser, c.Options.AuthPass)
	}

	return c.do(req)
}

// Returns true if the domain exists in the BasicAutDomains slice.
func (c *BasicClient) isBasicAuthDomain(domain string) bool {
	for _, authDomain := range c.Options.BasicAuthDomains {
		if authDomain == domain {
			return true
		}
	}

	return false
}

// Makes a GET request to an URL and returns the http response or an error.
func (c *BasicClient) Get(urlStr string) (*ClientResponse, error) {
	return c.request(http.MethodGet, urlStr)
}

// Makes a HEAD request to an URL and returns the http response or an error.
func (c *BasicClient) Head(urlStr string) (*ClientResponse, error) {
	return c.request(http.MethodHead, urlStr)
}

// do executes a request and returns its response and error.
// It sets the client's User-Agent as well as the BasicAuth details if they are available.
func (c *BasicClient) do(req *http.Request) (*ClientResponse, error) {
	cr := &ClientResponse{}

	start := time.Now()
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			// Time To First Byte in milliseconds
			cr.TTFB = int(math.Ceil(float64(time.Since(start) / time.Millisecond)))
		},
	}

	req.Header.Set("User-Agent", c.Options.UserAgent)
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	cr.Response = resp

	return cr, nil
}

// GetUA returns the user-agent set for this client.
func (c *BasicClient) GetUA() string {
	return c.Options.UserAgent
}
