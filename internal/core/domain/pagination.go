package domain

// PaginationRequest holds the requested page and limit
type PaginationRequest struct {
	Page   int
	Limit  int
	Search string
}

// PaginationMeta holds the standard metadata for pagination
type PaginationMeta struct {
	Page         int `json:"page"`
	Limit        int `json:"limit"`
	TotalRecords int `json:"total_records"`
	TotalPages   int `json:"total_pages"`
}

// PaginationResponse is a generic response for list endpoints
type PaginationResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}
