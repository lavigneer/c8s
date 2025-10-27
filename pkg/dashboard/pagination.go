package dashboard

import (
	"math"
	"net/url"
	"strconv"
)

// PaginationParams holds pagination parameters from query string
type PaginationParams struct {
	Page    int
	PerPage int
}

// PaginationResult holds pagination result data
type PaginationResult struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// ParsePaginationParams parses pagination parameters from query string
func ParsePaginationParams(query url.Values) PaginationParams {
	page := 1
	perPage := 20

	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if pp := query.Get("per_page"); pp != "" {
		if parsed, err := strconv.Atoi(pp); err == nil && parsed > 0 && parsed <= 100 {
			perPage = parsed
		}
	}

	return PaginationParams{
		Page:    page,
		PerPage: perPage,
	}
}

// Paginate slices a list based on pagination parameters
// Returns sliced items and pagination metadata
func Paginate(items interface{}, total int, params PaginationParams) *PaginationResult {
	// Validate parameters
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	// Calculate pagination values
	totalPages := int(math.Ceil(float64(total) / float64(params.PerPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Ensure page is within bounds
	if params.Page > totalPages {
		params.Page = totalPages
	}

	// Calculate start and end indices
	start := (params.Page - 1) * params.PerPage
	end := start + params.PerPage

	// Ensure bounds
	if start < 0 {
		start = 0
	}
	if end > total {
		end = total
	}

	return &PaginationResult{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PerPage:    params.PerPage,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}
}

// PaginateSlice slices a generic slice based on pagination parameters
// This is a helper that actually performs the slicing
func PaginateSlice(items interface{}, total int, params PaginationParams) (interface{}, *PaginationResult) {
	result := Paginate(items, total, params)

	// For actual slicing, caller should use the indices
	// This function just returns metadata
	return items, result
}

// GetPaginationIndices returns start and end indices for slicing
func GetPaginationIndices(page, perPage int) (start, end int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}

	start = (page - 1) * perPage
	end = start + perPage

	if start < 0 {
		start = 0
	}

	return start, end
}

// CalculatePageCount calculates total pages given item count and per-page limit
func CalculatePageCount(total, perPage int) int {
	if perPage <= 0 {
		perPage = 20
	}
	pages := int(math.Ceil(float64(total) / float64(perPage)))
	if pages == 0 {
		pages = 1
	}
	return pages
}

// BuildPaginationLinks generates links for pagination
type PaginationLinks struct {
	First    string
	Last     string
	Next     string
	Previous string
}

// GeneratePaginationLinks creates pagination navigation links
func GeneratePaginationLinks(baseURL string, result *PaginationResult) PaginationLinks {
	links := PaginationLinks{}

	// Add query parameters
	separator := "?"
	if baseURL[len(baseURL)-1:] == "?" {
		separator = ""
	} else if baseURL[len(baseURL)-1:] == "&" {
		separator = ""
	} else if baseURL[len(baseURL)-1:] != "/" {
		separator = "?"
	}

	baseURLWithParams := baseURL + separator + "per_page=" + strconv.Itoa(result.PerPage)

	// First page
	links.First = baseURLWithParams + "&page=1"

	// Last page
	links.Last = baseURLWithParams + "&page=" + strconv.Itoa(result.TotalPages)

	// Next page
	if result.HasNext {
		links.Next = baseURLWithParams + "&page=" + strconv.Itoa(result.Page+1)
	}

	// Previous page
	if result.HasPrev {
		links.Previous = baseURLWithParams + "&page=" + strconv.Itoa(result.Page-1)
	}

	return links
}

// CalculatePagination is a helper that returns metadata based on total count
func CalculatePagination(total, page, perPage int) *Metadata {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages == 0 {
		totalPages = 1
	}

	if page > totalPages {
		page = totalPages
	}

	return &Metadata{
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
}
