package utho

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const version = "1.49.0"

const BaseUrl = "https://api.utho.com/v2/"

var defaultHTTPClient = &http.Client{Timeout: time.Second * 5}

type Client interface {
	NewRequest(method, url string, body ...interface{}) (*http.Request, error)
	Do(req *http.Request, v interface{}) (*http.Response, error)

	Account() *AccountService
	ApiKey() *ApiKeyService
}

type service struct {
	client Client
}

type client struct {
	client  *http.Client
	baseURL *url.URL
	token   string

	account *AccountService
	apiKey  *ApiKeyService
}

// NewClient creates a new Clerk client.
// Because the token supplied will be used for all authenticated requests,
// the created client should not be used across different users
func NewClient(token string, options ...ClerkOption) (Client, error) {
	if token == "" {
		return nil, errors.New("you must provide an API token")
	}

	defaultBaseURL, err := toURLWithEndingSlash(BaseUrl)
	if err != nil {
		return nil, err
	}

	client := &client{
		client:  defaultHTTPClient,
		baseURL: defaultBaseURL,
		token:   token,
	}

	for _, option := range options {
		if err = option(client); err != nil {
			return nil, err
		}
	}

	commonService := &service{client: client}
	client.account = (*AccountService)(commonService)
	client.apiKey = (*ApiKeyService)(commonService)

	return client, nil
}

func toURLWithEndingSlash(u string) (*url.URL, error) {
	baseURL, err := url.Parse(u)
	if err != nil {
		return nil, err
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	return baseURL, err
}

// NewRequest creates an API request.
// A relative URL `url` can be specified which is resolved relative to the baseURL of the client.
// Relative URLs should be specified without a preceding slash.
// The `body` parameter can be used to pass a body to the request. If no body is required, the parameter can be omitted.
func (c *client) NewRequest(method, url string, body ...interface{}) (*http.Request, error) {
	fullUrl, err := c.baseURL.Parse(url)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if len(body) > 0 && body[0] != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body[0])
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, fullUrl.String(), buf)
	if err != nil {
		return nil, err
	}

	// Add custom header with the current SDK version
	req.Header.Set("X-Clerk-SDK", fmt.Sprintf("go/%s", version))

	return req, nil
}

// Do will send the given request using the client `c` on which it is called.
// If the response contains a body, it will be unmarshalled in `v`.
func (c *client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = checkForErrors(resp)
	if err != nil {
		return resp, err
	}

	if resp.Body != nil && v != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return resp, err
		}

		err = json.Unmarshal(body, &v)
		if err != nil {
			return resp, err
		}
	}

	return resp, nil
}

func checkForErrors(resp *http.Response) error {
	if c := resp.StatusCode; c >= 200 && c < 400 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: resp}

	data, err := io.ReadAll(resp.Body)
	if err == nil && data != nil {
		// it's ok if we cannot unmarshal to Clerk's error response
		_ = json.Unmarshal(data, errorResponse)
	}

	return errorResponse
}

func (c *client) Account() *AccountService {
	return c.account
}

func (c *client) ApiKey() *ApiKeyService {
	return c.apiKey
}