package types

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Success bool   `json:"success"`
	Details any    `json:"details,omitempty"`
}

// SuccessResponse represents a standard success response
type SuccessResponse struct {
	Message string `json:"message,omitempty"`
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalPages int `json:"total_pages"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}
