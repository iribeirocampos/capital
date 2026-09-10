package capital

import (
	"encoding/json"
	"fmt"
)

// APIError represents an error response returned by the Capital.com API.
// Use errors.As to recover one from an error returned by a Client method.
type APIError struct {
	// StatusCode is the HTTP status code of the response.
	StatusCode int
	// Code is the Capital.com error code (the "errorCode" field), when the
	// API returned one.
	Code string
	// Body is the raw response body, for cases where Code could not be
	// parsed out of it.
	Body string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("capital: api error: status %d, code %q", e.StatusCode, e.Code)
	}
	return fmt.Sprintf("capital: api error: status %d: %s", e.StatusCode, e.Body)
}

type apiErrorBody struct {
	ErrorCode string `json:"errorCode"`
}

func parseAPIError(statusCode int, body []byte) *APIError {
	apiErr := &APIError{StatusCode: statusCode, Body: string(body)}
	var parsed apiErrorBody
	if err := json.Unmarshal(body, &parsed); err == nil {
		apiErr.Code = parsed.ErrorCode
	}
	return apiErr
}
