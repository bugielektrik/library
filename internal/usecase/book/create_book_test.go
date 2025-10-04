package book

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"

	"library-service/internal/domain/book"
	"library-service/internal/adapters/repository/mock"
	"library-service/pkg/errors"
)

func TestCreateBookUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		request       CreateBookRequest
		setupMocks    func(*mock.MockRepository, *mock.MockCache)
		expectedError error
		validateFunc  func(*testing.T, CreateBookResponse)
	}{
		{
			name: "successful book creation",
			request: CreateBookRequest{
				Name:    "The Go Programming Language",
				Genre:   "Technology",
				ISBN:    "978-0134190440",
				Authors: []string{"author-id-1"},
			},
			setupMocks: func(repo *mock.MockRepository, cache *mock.MockCache) {
				// Expect check for existing book (none found)
				repo.EXPECT().
					Get(testifymock.Anything, "978-0134190440").
					Return(book.Entity{}, assert.AnError).
					Once()

				// Expect book creation
				repo.EXPECT().
					Add(testifymock.Anything, testifymock.MatchedBy(func(b book.Entity) bool {
						return *b.Name == "The Go Programming Language" &&
							*b.Genre == "Technology" &&
							*b.ISBN == "978-0134190440"
					})).
					Return("book-123", nil).
					Once()

				// Expect cache update
				cache.EXPECT().
					Set(testifymock.Anything, "book-123", testifymock.Anything).
					Return(nil).
					Once()
			},
			expectedError: nil,
			validateFunc: func(t *testing.T, resp CreateBookResponse) {
				assert.Equal(t, "book-123", resp.ID)
				assert.Equal(t, "The Go Programming Language", resp.Name)
				assert.Equal(t, "Technology", resp.Genre)
				assert.Equal(t, "978-0134190440", resp.ISBN)
			},
		},
		{
			name: "validation error - empty name",
			request: CreateBookRequest{
				Name:    "",
				Genre:   "Technology",
				ISBN:    "978-0134190440",
				Authors: []string{"author-id-1"},
			},
			setupMocks: func(repo *mock.MockRepository, cache *mock.MockCache) {
				// No mock calls expected as validation fails first
			},
			expectedError: errors.ErrInvalidBookData,
		},
		{
			name: "validation error - empty ISBN",
			request: CreateBookRequest{
				Name:    "The Go Programming Language",
				Genre:   "Technology",
				ISBN:    "",
				Authors: []string{"author-id-1"},
			},
			setupMocks: func(repo *mock.MockRepository, cache *mock.MockCache) {
				// No mock calls expected
			},
			expectedError: errors.ErrInvalidISBN,
		},
		{
			name: "validation error - no authors",
			request: CreateBookRequest{
				Name:    "The Go Programming Language",
				Genre:   "Technology",
				ISBN:    "978-0134190440",
				Authors: []string{},
			},
			setupMocks: func(repo *mock.MockRepository, cache *mock.MockCache) {
				// No mock calls expected
			},
			expectedError: errors.ErrInvalidBookData,
		},
		{
			name: "book already exists",
			request: CreateBookRequest{
				Name:    "The Go Programming Language",
				Genre:   "Technology",
				ISBN:    "978-0134190440",
				Authors: []string{"author-id-1"},
			},
			setupMocks: func(repo *mock.MockRepository, cache *mock.MockCache) {
				// Book with ISBN already exists
				existingBook := book.Entity{
					ID: "existing-book-id",
				}
				repo.EXPECT().
					Get(testifymock.Anything, "978-0134190440").
					Return(existingBook, nil).
					Once()
			},
			expectedError: errors.ErrBookAlreadyExists,
		},
		{
			name: "repository error during creation",
			request: CreateBookRequest{
				Name:    "The Go Programming Language",
				Genre:   "Technology",
				ISBN:    "978-0134190440",
				Authors: []string{"author-id-1"},
			},
			setupMocks: func(repo *mock.MockRepository, cache *mock.MockCache) {
				// Check for existing book (none)
				repo.EXPECT().
					Get(testifymock.Anything, "978-0134190440").
					Return(book.Entity{}, assert.AnError).
					Once()

				// Repository error during creation
				repo.EXPECT().
					Add(testifymock.Anything, testifymock.Anything).
					Return("", assert.AnError).
					Once()
			},
			expectedError: errors.ErrDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := mock.NewMockRepository(t)
			mockCache := mock.NewMockCache(t)
			tt.setupMocks(mockRepo, mockCache)

			// Create domain service and use case
			bookService := book.NewService()
			uc := NewCreateBookUseCase(mockRepo, mockCache, bookService)

			// Execute
			result, err := uc.Execute(context.Background(), tt.request)

			// Assert error
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}
