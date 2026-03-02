package models

import "fmt"

// APIError represents an error response from the Dune API
type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"error"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}
