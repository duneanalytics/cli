package sim

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://api.sim.dune.com"

// SimClient is a lightweight HTTP client for the Sim API.
type SimClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewSimClient creates a new Sim API client with the given API key.
func NewSimClient(apiKey string) *SimClient {
	return &SimClient{
		baseURL: defaultBaseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewBareSimClient creates a Sim API client without authentication.
// Use this for public endpoints that don't require an API key.
func NewBareSimClient() *SimClient {
	return &SimClient{
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get performs a GET request to the Sim API and returns the raw JSON response body.
// The path should include the leading slash (e.g. "/v1/evm/supported-chains").
// Query parameters are appended from params.
func (c *SimClient) Get(ctx context.Context, path string, params url.Values) ([]byte, error) {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	if params != nil {
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if c.apiKey != "" {
		req.Header.Set("X-Sim-Api-Key", c.apiKey)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, httpError(resp.StatusCode, body)
	}

	return body, nil
}

// httpError returns a user-friendly error for HTTP error status codes.
func httpError(status int, body []byte) error {
	// Try to extract a message from the JSON error response.
	var errResp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	msg := ""
	if json.Unmarshal(body, &errResp) == nil {
		if errResp.Error != "" {
			msg = errResp.Error
		} else if errResp.Message != "" {
			msg = errResp.Message
		}
	}

	switch status {
	case http.StatusBadRequest:
		if msg != "" {
			return fmt.Errorf("bad request: %s", msg)
		}
		return fmt.Errorf("bad request")
	case http.StatusUnauthorized:
		return fmt.Errorf("authentication failed: check your Sim API key")
	case http.StatusNotFound:
		if msg != "" {
			return fmt.Errorf("not found: %s", msg)
		}
		return fmt.Errorf("not found")
	case http.StatusTooManyRequests:
		return fmt.Errorf("rate limit exceeded: try again later")
	default:
		if status >= 500 {
			return fmt.Errorf("Sim API server error (HTTP %d): try again later", status)
		}
		if msg != "" {
			return fmt.Errorf("Sim API error (HTTP %d): %s", status, msg)
		}
		return fmt.Errorf("Sim API error (HTTP %d)", status)
	}
}
