package pagination_test

import (
	"fmt"
	"library-service/internal/pkg/pagination"
)

// Example demonstrates basic pagination usage
func Example() {
	// Create paginator for page 1, 10 items per page
	p := pagination.NewPaginator(1, 10)

	fmt.Println("Page:", p.Page)
	fmt.Println("Page size:", p.PageSize)
	fmt.Println("Offset:", p.Offset())

	// Output:
	// Page: 1
	// Page size: 10
	// Offset: 0
}

// ExampleNewPaginator demonstrates creating a paginator
func ExampleNewPaginator() {
	// Valid pagination
	p1 := pagination.NewPaginator(2, 20)
	fmt.Printf("Page %d, Size %d\n", p1.Page, p1.PageSize)

	// Auto-corrects invalid page (< 1)
	p2 := pagination.NewPaginator(0, 10)
	fmt.Printf("Invalid page corrected to: %d\n", p2.Page)

	// Auto-corrects invalid page size (< 1)
	p3 := pagination.NewPaginator(1, 0)
	fmt.Printf("Invalid size corrected to: %d\n", p3.PageSize)

	// Limits max page size to 100
	p4 := pagination.NewPaginator(1, 200)
	fmt.Printf("Large size limited to: %d\n", p4.PageSize)

	// Output:
	// Page 2, Size 20
	// Invalid page corrected to: 1
	// Invalid size corrected to: 10
	// Large size limited to: 100
}

// ExamplePaginator_Offset demonstrates offset calculation
func ExamplePaginator_Offset() {
	// Page 1: offset 0
	p1 := pagination.NewPaginator(1, 10)
	fmt.Printf("Page 1, offset: %d\n", p1.Offset())

	// Page 2: offset 10
	p2 := pagination.NewPaginator(2, 10)
	fmt.Printf("Page 2, offset: %d\n", p2.Offset())

	// Page 3: offset 20
	p3 := pagination.NewPaginator(3, 10)
	fmt.Printf("Page 3, offset: %d\n", p3.Offset())

	// Output:
	// Page 1, offset: 0
	// Page 2, offset: 10
	// Page 3, offset: 20
}

// ExamplePaginator_BuildPage demonstrates building paginated response
func ExamplePaginator_BuildPage() {
	p := pagination.NewPaginator(1, 10)

	// Simulate 25 total items
	items := []string{"item1", "item2", "item3"}
	page := p.BuildPage(items, 25)

	fmt.Println("Total:", page.Total)
	fmt.Println("Total pages:", page.TotalPages)
	fmt.Println("Current page:", page.Page)
	fmt.Println("Has next:", page.HasNext)
	fmt.Println("Has prev:", page.HasPrev)

	// Output:
	// Total: 25
	// Total pages: 3
	// Current page: 1
	// Has next: true
	// Has prev: false
}

// ExamplePaginator_BuildPage_lastPage demonstrates last page handling
func ExamplePaginator_BuildPage_lastPage() {
	p := pagination.NewPaginator(3, 10)

	// Page 3 of 3 (25 total items)
	items := []string{"item21", "item22", "item23", "item24", "item25"}
	page := p.BuildPage(items, 25)

	fmt.Println("Current page:", page.Page)
	fmt.Println("Total pages:", page.TotalPages)
	fmt.Println("Has next:", page.HasNext)
	fmt.Println("Has prev:", page.HasPrev)

	// Output:
	// Current page: 3
	// Total pages: 3
	// Has next: false
	// Has prev: true
}

// Example_databaseQuery demonstrates using pagination with database queries
func Example_databaseQuery() {
	// Simulate paginated database query
	p := pagination.NewPaginator(2, 10) // Page 2, 10 items

	// In real code:
	// offset := p.Offset()  // 10
	// limit := p.Limit()    // 10
	// items := db.Query("SELECT * FROM books LIMIT ? OFFSET ?", limit, offset)
	// total := db.Count("SELECT COUNT(*) FROM books")

	// Simulated result
	offset := p.Offset()
	limit := p.Limit()

	fmt.Printf("SQL: SELECT * FROM books LIMIT %d OFFSET %d\n", limit, offset)
	fmt.Println("Returns items 11-20")

	// Output:
	// SQL: SELECT * FROM books LIMIT 10 OFFSET 10
	// Returns items 11-20
}

// Example_paginationWorkflow demonstrates complete pagination workflow
func Example_paginationWorkflow() {
	// 1. Parse request (page, pageSize from query params)
	page, pageSize := 1, 10

	// 2. Create paginator
	p := pagination.NewPaginator(page, pageSize)

	// 3. Query database with offset/limit
	// items := db.Query(p.Offset(), p.Limit())
	// total := db.Count()

	// Simulated data
	items := []map[string]string{
		{"id": "1", "title": "Book 1"},
		{"id": "2", "title": "Book 2"},
	}
	total := 25

	// 4. Build paginated response
	response := p.BuildPage(items, total)

	fmt.Printf("Showing page %d of %d\n", response.Page, response.TotalPages)
	fmt.Printf("Items on this page: %d\n", len(items))
	fmt.Printf("Total items: %d\n", response.Total)

	// Output:
	// Showing page 1 of 3
	// Items on this page: 2
	// Total items: 25
}

// Example_edgeCases demonstrates edge case handling
func Example_edgeCases() {
	// Empty results
	p1 := pagination.NewPaginator(1, 10)
	page1 := p1.BuildPage([]string{}, 0)
	fmt.Printf("Empty: Total pages = %d\n", page1.TotalPages)

	// Single item
	p2 := pagination.NewPaginator(1, 10)
	page2 := p2.BuildPage([]string{"item1"}, 1)
	fmt.Printf("Single item: Total pages = %d\n", page2.TotalPages)

	// Exact page boundary (10 items, 10 per page)
	p3 := pagination.NewPaginator(1, 10)
	page3 := p3.BuildPage([]string{}, 10)
	fmt.Printf("Exact boundary: Total pages = %d\n", page3.TotalPages)

	// One over boundary (11 items, 10 per page)
	p4 := pagination.NewPaginator(1, 10)
	page4 := p4.BuildPage([]string{}, 11)
	fmt.Printf("One over: Total pages = %d\n", page4.TotalPages)

	// Output:
	// Empty: Total pages = 1
	// Single item: Total pages = 1
	// Exact boundary: Total pages = 1
	// One over: Total pages = 2
}
