package pagination

import (
	"math"
	"strconv"
)

// DefaultParams contains the default pagination parameters
type DefaultParams struct {
	DefaultPage int
	DefaultSize int
	MaxSize     int
}

// Parse parses and validates pagination parameters from query parameters
func Parse(page, size string, defaults DefaultParams) (int, int) {
	// Parse page number
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		pageNum = defaults.DefaultPage
	}

	// Parse page size
	pageSize, err := strconv.Atoi(size)
	if err != nil || pageSize < 1 {
		pageSize = defaults.DefaultSize
	}

	// Limit maximum page size
	if pageSize > defaults.MaxSize {
		pageSize = defaults.MaxSize
	}

	return pageNum, pageSize
}

// CalculateTotalPages calculates the total number of pages based on total items and page size
func CalculateTotalPages(totalItems, pageSize int) int {
	return int(math.Ceil(float64(totalItems) / float64(pageSize)))
}
