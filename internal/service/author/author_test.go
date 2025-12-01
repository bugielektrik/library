package author

import (
	"context"
	"errors"
	"library-service/internal/domain/author"
	"library-service/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthorService_GetAuthor_Success(t *testing.T) {
	mockRepo := mocks.NewMockAuthorRepository(t)
	mockCache := mocks.NewMockAuthorCache(t)
	ctx := context.Background()

	authorID := uuid.New().String()
	fullName := "Абай Кунанбаев"
	pseudonym := "Какитай"
	specialty := "Поэт"

	expectedAuthor := author.Entity{
		ID:        authorID,
		FullName:  &fullName,
		Pseudonym: &pseudonym,
		Specialty: &specialty,
	}

	mockCache.EXPECT().Get(mock.Anything, mock.Anything).Return(author.Entity{}, errors.New("cache miss")).Maybe()
	mockCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	mockRepo.EXPECT().Get(ctx, authorID).Return(expectedAuthor, nil).Once()

	service := NewAuthorService(mockRepo, mockCache)

	result, err := service.GetAuthor(ctx, authorID)

	assert.NoError(t, err)
	assert.Equal(t, expectedAuthor.ID, result.ID)
	assert.Equal(t, *expectedAuthor.FullName, result.FullName)
	assert.Equal(t, *expectedAuthor.Pseudonym, result.Pseudonym)
}
func TestAuthorService_GetAuthor_Fail(t *testing.T) {
	mockRepo := mocks.NewMockAuthorRepository(t)
	mockCache := mocks.NewMockAuthorCache(t)
	ctx := context.Background()

	authorID := uuid.New().String()

	mockCache.EXPECT().Get(mock.Anything, mock.Anything).Return(author.Entity{}, errors.New("cache miss")).Maybe()
	mockCache.EXPECT().Set(mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

	mockRepo.EXPECT().Get(ctx, authorID).Return(author.Entity{}, errors.New("")).Once()

	service := NewAuthorService(mockRepo, mockCache)

	result, err := service.GetAuthor(ctx, authorID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "")

	assert.Empty(t, result.ID)
	assert.Empty(t, result.FullName)
	assert.Empty(t, result.Pseudonym)
	assert.Empty(t, result.Specialty)
}

func TestAuthorService_AddAuthor_Success(t *testing.T) {

}
func TestAuthorService_AddAuthor_Fail(t *testing.T) {

}

func TestAuthorService_UpdateAuthor_Success(t *testing.T) {

}
func TestAuthorService_UpdateAuthor_Fail(t *testing.T) {

}

func TestAuthorService_ListAuthor_Success(t *testing.T) {

}
func TestAuthorService_ListAuthor_Fail(t *testing.T) {

}
