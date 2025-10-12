package httputil_test

import (
	"bytes"
	"fmt"
	httputil2 "library-service/internal/pkg/httputil"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"
)

// Example demonstrates basic HTTP utility usage
func Example() {
	// Check if status code is successful
	isOK := httputil2.IsSuccess(200)
	fmt.Println("200 is success:", isOK)

	// Output:
	// 200 is success: true
}

// ExampleIsSuccess demonstrates success status checking
func ExampleIsSuccess() {
	fmt.Println("200:", httputil2.IsSuccess(200))
	fmt.Println("201:", httputil2.IsSuccess(201))
	fmt.Println("204:", httputil2.IsSuccess(204))
	fmt.Println("300:", httputil2.IsSuccess(300))
	fmt.Println("400:", httputil2.IsSuccess(400))

	// Output:
	// 200: true
	// 201: true
	// 204: true
	// 300: false
	// 400: false
}

// ExampleIsClientError demonstrates client error status checking
func ExampleIsClientError() {
	fmt.Println("400:", httputil2.IsClientError(400))
	fmt.Println("404:", httputil2.IsClientError(404))
	fmt.Println("422:", httputil2.IsClientError(422))
	fmt.Println("200:", httputil2.IsClientError(200))
	fmt.Println("500:", httputil2.IsClientError(500))

	// Output:
	// 400: true
	// 404: true
	// 422: true
	// 200: false
	// 500: false
}

// ExampleIsServerError demonstrates server error status checking
func ExampleIsServerError() {
	fmt.Println("500:", httputil2.IsServerError(500))
	fmt.Println("502:", httputil2.IsServerError(502))
	fmt.Println("503:", httputil2.IsServerError(503))
	fmt.Println("400:", httputil2.IsServerError(400))
	fmt.Println("200:", httputil2.IsServerError(200))

	// Output:
	// 500: true
	// 502: true
	// 503: true
	// 400: false
	// 200: false
}

// ExampleIsRedirect demonstrates redirect status checking
func ExampleIsRedirect() {
	fmt.Println("301:", httputil2.IsRedirect(301))
	fmt.Println("302:", httputil2.IsRedirect(302))
	fmt.Println("307:", httputil2.IsRedirect(307))
	fmt.Println("200:", httputil2.IsRedirect(200))
	fmt.Println("404:", httputil2.IsRedirect(404))

	// Output:
	// 301: true
	// 302: true
	// 307: true
	// 200: false
	// 404: false
}

// ExampleIsError demonstrates error status checking (4xx or 5xx)
func ExampleIsError() {
	fmt.Println("400:", httputil2.IsError(400))
	fmt.Println("404:", httputil2.IsError(404))
	fmt.Println("500:", httputil2.IsError(500))
	fmt.Println("200:", httputil2.IsError(200))
	fmt.Println("301:", httputil2.IsError(301))

	// Output:
	// 400: true
	// 404: true
	// 500: true
	// 200: false
	// 301: false
}

// ExampleDecodeJSON demonstrates JSON request body decoding
func ExampleDecodeJSON() {
	// Create test request with JSON body
	body := `{"name": "Clean Code", "author": "Robert Martin"}`
	req := httptest.NewRequest("POST", "/books", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// Define target struct
	type BookRequest struct {
		Name   string `json:"name"`
		Author string `json:"author"`
	}

	var book BookRequest
	err := httputil2.DecodeJSON(req, &book)

	fmt.Println("Error:", err)
	fmt.Println("Name:", book.Name)
	fmt.Println("Author:", book.Author)

	// Output:
	// Error: <nil>
	// Name: Clean Code
	// Author: Robert Martin
}

// ExampleGetURLParam demonstrates URL parameter extraction
func ExampleGetURLParam() {
	// Create chi router with URL parameter
	r := chi.NewRouter()

	var capturedID string
	var capturedErr error

	r.Get("/books/{id}", func(w http.ResponseWriter, req *http.Request) {
		capturedID, capturedErr = httputil2.GetURLParam(req, "id")
	})

	// Test with valid parameter
	req := httptest.NewRequest("GET", "/books/book-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	fmt.Println("ID:", capturedID)
	fmt.Println("Error:", capturedErr)

	// Output:
	// ID: book-123
	// Error: <nil>
}

// ExampleGetURLParam_missing demonstrates missing parameter handling
func ExampleGetURLParam_missing() {
	// Create request without parameter
	req := httptest.NewRequest("GET", "/books", nil)

	// Try to get non-existent parameter
	_, err := httputil2.GetURLParam(req, "id")

	fmt.Println("Has error:", err != nil)

	// Output:
	// Has error: true
}

// Example_statusCodeCategories demonstrates categorizing HTTP status codes
func Example_statusCodeCategories() {
	codes := []int{200, 301, 400, 404, 500, 503}

	for _, code := range codes {
		category := "Unknown"
		switch {
		case httputil2.IsSuccess(code):
			category = "Success"
		case httputil2.IsRedirect(code):
			category = "Redirect"
		case httputil2.IsClientError(code):
			category = "Client Error"
		case httputil2.IsServerError(code):
			category = "Server Error"
		}
		fmt.Printf("%d: %s\n", code, category)
	}

	// Output:
	// 200: Success
	// 301: Redirect
	// 400: Client Error
	// 404: Client Error
	// 500: Server Error
	// 503: Server Error
}

// Example_handlerPattern demonstrates common handler patterns using httputil
func Example_handlerPattern() {
	type CreateBookRequest struct {
		Title string `json:"title"`
		ISBN  string `json:"isbn"`
	}

	// Simulate handler logic
	body := `{"title": "Domain-Driven Design", "isbn": "978-0321125217"}`
	req := httptest.NewRequest("POST", "/api/books", bytes.NewBufferString(body))

	var request CreateBookRequest
	if err := httputil2.DecodeJSON(req, &request); err != nil {
		fmt.Println("Decode error:", err)
		return
	}

	// Validate
	if request.Title == "" {
		fmt.Println("Title is required")
		return
	}

	fmt.Println("Title:", request.Title)
	fmt.Println("ISBN:", request.ISBN)

	// Output:
	// Title: Domain-Driven Design
	// ISBN: 978-0321125217
}
