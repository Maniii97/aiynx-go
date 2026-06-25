package models

// ErrorDetail carries a human-readable message, a machine-readable code,
// and the request ID for tracing — consistent across every handler.
type ErrorDetail struct {
	Message   string `json:"message"`
	Code      string `json:"code"`
	RequestID string `json:"request_id"`
}

// ErrorResponse is the standard error envelope returned by all endpoints.
//
// Error codes used across the application:
//
//	BAD_REQUEST         – 400  parse / bad JSON
//	VALIDATION_ERROR    – 422  validation failure
//	INVALID_PARAM       – 400  invalid :id param
//	CONFLICT            – 409  email already exists
//	INVALID_CREDENTIALS – 401  wrong credentials
//	UNAUTHORIZED        – 401  missing / invalid token
//	FORBIDDEN           – 403  insufficient role
//	NOT_FOUND           – 404  resource not found
//	INTERNAL_ERROR      – 500  unexpected / DB error
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
