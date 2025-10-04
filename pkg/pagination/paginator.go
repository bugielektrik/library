package pagination

// Page represents a paginated response
type Page struct {
	Items      interface{} `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// Paginator handles pagination logic
type Paginator struct {
	Page     int
	PageSize int
}

// NewPaginator creates a new Paginator
func NewPaginator(page, pageSize int) *Paginator {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &Paginator{
		Page:     page,
		PageSize: pageSize,
	}
}

// Offset calculates the offset for database queries
func (p *Paginator) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Limit returns the page size limit
func (p *Paginator) Limit() int {
	return p.PageSize
}

// BuildPage creates a Page response
func (p *Paginator) BuildPage(items interface{}, total int) Page {
	totalPages := (total + p.PageSize - 1) / p.PageSize
	if totalPages < 1 {
		totalPages = 1
	}

	return Page{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
		HasNext:    p.Page < totalPages,
		HasPrev:    p.Page > 1,
	}
}
