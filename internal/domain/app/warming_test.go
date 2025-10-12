package app

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
)

// mockBookRepository for testing
type mockBookRepository struct {
	books []book.Book
}

func (m *mockBookRepository) List(ctx context.Context) ([]book.Book, error) {
	return m.books, nil
}

func (m *mockBookRepository) Add(ctx context.Context, data book.Book) (string, error) {
	return "", nil
}

func (m *mockBookRepository) Get(ctx context.Context, id string) (book.Book, error) {
	return book.Book{}, nil
}

func (m *mockBookRepository) Update(ctx context.Context, id string, data book.Book) error {
	return nil
}

func (m *mockBookRepository) Delete(ctx context.Context, id string) error {
	return nil
}

// mockAuthorRepository for testing
type mockAuthorRepository struct {
	authors []author.Author
}

func (m *mockAuthorRepository) List(ctx context.Context) ([]author.Author, error) {
	return m.authors, nil
}

func (m *mockAuthorRepository) Add(ctx context.Context, data author.Author) (string, error) {
	return "", nil
}

func (m *mockAuthorRepository) Get(ctx context.Context, id string) (author.Author, error) {
	return author.Author{}, nil
}

func (m *mockAuthorRepository) Update(ctx context.Context, id string, data author.Author) error {
	return nil
}

func (m *mockAuthorRepository) Delete(ctx context.Context, id string) error {
	return nil
}

// mockBookCache for testing with thread-safe access
type mockBookCache struct {
	mu    sync.RWMutex
	books map[string]book.Book
}

func (m *mockBookCache) Get(ctx context.Context, id string) (book.Book, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if b, ok := m.books[id]; ok {
		return b, nil
	}
	return book.Book{}, nil
}

func (m *mockBookCache) Set(ctx context.Context, id string, b book.Book) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.books == nil {
		m.books = make(map[string]book.Book)
	}
	m.books[id] = b
	return nil
}

func (m *mockBookCache) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.books)
}

// mockAuthorCache for testing with thread-safe access
type mockAuthorCache struct {
	mu      sync.RWMutex
	authors map[string]author.Author
}

func (m *mockAuthorCache) Get(ctx context.Context, id string) (author.Author, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if a, ok := m.authors[id]; ok {
		return a, nil
	}
	return author.Author{}, nil
}

func (m *mockAuthorCache) Set(ctx context.Context, id string, a author.Author) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.authors == nil {
		m.authors = make(map[string]author.Author)
	}
	m.authors[id] = a
	return nil
}

func (m *mockAuthorCache) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.authors)
}

func TestWarmCaches(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		books          []book.Book
		authors        []author.Author
		config         WarmingConfig
		expectedBooks  int
		expectedAuthor int
	}{
		{
			name: "warm all books and authors",
			books: []book.Book{
				{ID: "book-1", Name: stringPtr("Book 1")},
				{ID: "book-2", Name: stringPtr("Book 2")},
				{ID: "book-3", Name: stringPtr("Book 3")},
			},
			authors: []author.Author{
				{ID: "author-1", FullName: stringPtr("Author 1")},
				{ID: "author-2", FullName: stringPtr("Author 2")},
			},
			config: WarmingConfig{
				Enabled:            true,
				PopularBookLimit:   10,
				PopularAuthorLimit: 10,
				Timeout:            5 * time.Second,
				Logger:             logger,
			},
			expectedBooks:  3,
			expectedAuthor: 2,
		},
		{
			name: "warm limited books",
			books: []book.Book{
				{ID: "book-1", Name: stringPtr("Book 1")},
				{ID: "book-2", Name: stringPtr("Book 2")},
				{ID: "book-3", Name: stringPtr("Book 3")},
				{ID: "book-4", Name: stringPtr("Book 4")},
				{ID: "book-5", Name: stringPtr("Book 5")},
			},
			authors: []author.Author{
				{ID: "author-1", FullName: stringPtr("Author 1")},
				{ID: "author-2", FullName: stringPtr("Author 2")},
				{ID: "author-3", FullName: stringPtr("Author 3")},
			},
			config: WarmingConfig{
				Enabled:            true,
				PopularBookLimit:   2,
				PopularAuthorLimit: 1,
				Timeout:            5 * time.Second,
				Logger:             logger,
			},
			expectedBooks:  2,
			expectedAuthor: 1,
		},
		{
			name:    "warming disabled",
			books:   []book.Book{{ID: "book-1", Name: stringPtr("Book 1")}},
			authors: []author.Author{{ID: "author-1", FullName: stringPtr("Author 1")}},
			config: WarmingConfig{
				Enabled:            false,
				PopularBookLimit:   10,
				PopularAuthorLimit: 10,
				Timeout:            5 * time.Second,
				Logger:             logger,
			},
			expectedBooks:  0,
			expectedAuthor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			bookRepo := &mockBookRepository{books: tt.books}
			authorRepo := &mockAuthorRepository{authors: tt.authors}
			bookCache := &mockBookCache{}
			authorCache := &mockAuthorCache{}

			caches := &Caches{
				Book:   bookCache,
				Author: authorCache,
				dependencies: Dependencies{
					Repositories: &Repositories{
						Book:   bookRepo,
						Author: authorRepo,
					},
				},
			}

			// Execute
			ctx := context.Background()
			err := WarmCaches(ctx, caches, tt.config)

			// Verify
			if err != nil {
				t.Errorf("WarmCaches() error = %v", err)
				return
			}

			if bookCache.Len() != tt.expectedBooks {
				t.Errorf("Expected %d books in cache, got %d", tt.expectedBooks, bookCache.Len())
			}

			if authorCache.Len() != tt.expectedAuthor {
				t.Errorf("Expected %d authors in cache, got %d", tt.expectedAuthor, authorCache.Len())
			}
		})
	}
}

// stringPtr is a helper to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestWarmCachesAsync(t *testing.T) {
	logger := zap.NewNop()

	// Setup
	books := []book.Book{
		{ID: "book-1", Name: stringPtr("Book 1")},
		{ID: "book-2", Name: stringPtr("Book 2")},
	}
	authors := []author.Author{
		{ID: "author-1", FullName: stringPtr("Author 1")},
	}

	bookRepo := &mockBookRepository{books: books}
	authorRepo := &mockAuthorRepository{authors: authors}
	bookCache := &mockBookCache{}
	authorCache := &mockAuthorCache{}

	caches := &Caches{
		Book:   bookCache,
		Author: authorCache,
		dependencies: Dependencies{
			Repositories: &Repositories{
				Book:   bookRepo,
				Author: authorRepo,
			},
		},
	}

	config := WarmingConfig{
		Enabled:            true,
		PopularBookLimit:   10,
		PopularAuthorLimit: 10,
		Timeout:            5 * time.Second,
		Logger:             logger,
	}

	// Execute async warming
	ctx := context.Background()
	WarmCachesAsync(ctx, caches, config)

	// Wait for warming to complete (poll with timeout)
	timeout := time.After(2 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for cache warming to complete")
		case <-ticker.C:
			if bookCache.Len() == 2 && authorCache.Len() == 1 {
				// Warming complete
				return
			}
		}
	}
}

func TestDefaultWarmingConfig(t *testing.T) {
	logger := zap.NewNop()

	config := DefaultWarmingConfig(logger)

	if !config.Enabled {
		t.Error("Expected warming to be enabled by default")
	}

	if config.PopularBookLimit != 50 {
		t.Errorf("Expected default book limit of 50, got %d", config.PopularBookLimit)
	}

	if config.PopularAuthorLimit != 20 {
		t.Errorf("Expected default author limit of 20, got %d", config.PopularAuthorLimit)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout of 30s, got %v", config.Timeout)
	}

	if config.Logger == nil {
		t.Error("Expected logger to be set")
	}
}
