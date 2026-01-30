package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"library-service/internal/domain/author"
	"library-service/internal/service/interfaces/mocks"
	"library-service/pkg/store"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthorHandler_Get_Success(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	authorID := "123e4567-e89b-12d3-a456-426614174000"
	expectedResponse := author.Response{
		ID:        authorID,
		FullName:  "Абай Кунанбаев",
		Pseudonym: "Какитай",
		Specialty: "Поэт",
	}

	mockService.EXPECT().
		GetAuthor(mock.Anything, authorID).
		Return(expectedResponse, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/authors/"+authorID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", authorID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.get(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response author.Response
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.Equal(t, expectedResponse.FullName, response.FullName)
	assert.Equal(t, expectedResponse.Pseudonym, response.Pseudonym)
	assert.Equal(t, expectedResponse.Specialty, response.Specialty)
}

func TestAuthorHandler_Get_NotFound(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	authorID := "non-existent-id"

	mockService.EXPECT().
		GetAuthor(mock.Anything, authorID).
		Return(author.Response{}, store.ErrorNotFound).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/authors/"+authorID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", authorID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.get(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAuthorHandler_Get_InternalError(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	authorID := "123e4567-e89b-12d3-a456-426614174000"

	mockService.EXPECT().
		GetAuthor(mock.Anything, authorID).
		Return(author.Response{}, errors.New("database connection failed")).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/authors/"+authorID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", authorID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	// Act
	handler.get(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAuthorHandler_List_Success(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	expectedAuthors := []author.Response{
		{
			ID:        "id1",
			FullName:  "Автор 1",
			Pseudonym: "Псевдоним 1",
			Specialty: "Специальность 1",
		},
		{
			ID:        "id2",
			FullName:  "Автор 2",
			Pseudonym: "Псевдоним 2",
			Specialty: "Специальность 2",
		},
	}

	mockService.EXPECT().
		ListAuthors(mock.Anything).
		Return(expectedAuthors, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/authors", nil)
	w := httptest.NewRecorder()

	handler.list(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []author.Response
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Len(t, response, 2)
	assert.Equal(t, expectedAuthors[0].ID, response[0].ID)
	assert.Equal(t, expectedAuthors[1].ID, response[1].ID)
}

func TestAuthorHandler_Add_Success(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	requestBody := author.Request{
		FullName:  "Новый автор",
		Pseudonym: "Новый псевдоним",
		Specialty: "Новая специальность",
	}

	expectedResponse := author.Response{
		ID:        "new-id",
		FullName:  requestBody.FullName,
		Pseudonym: requestBody.Pseudonym,
		Specialty: requestBody.Specialty,
	}

	mockService.EXPECT().
		AddAuthor(mock.Anything, requestBody).
		Return(expectedResponse, nil).
		Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/authors", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.add(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response author.Response
	err := json.NewDecoder(w.Body).Decode(&response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.Equal(t, expectedResponse.FullName, response.FullName)
}

func TestAuthorHandler_Add_BadRequest_InvalidJSON(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	invalidJSON := []byte(`{"fullName": "test"`)
	req := httptest.NewRequest(http.MethodPost, "/authors", bytes.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.add(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthorHandler_Add_BadRequest_ValidationError(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	requestBody := author.Request{
		FullName:  "",
		Pseudonym: "Псевдоним",
		Specialty: "Специальность",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/authors", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.add(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAuthorHandler_Update_Success(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	authorID := "123e4567-e89b-12d3-a456-426614174000"
	requestBody := author.Request{
		FullName:  "Обновленный автор",
		Pseudonym: "Обновленный псевдоним",
		Specialty: "Обновленная специальность",
	}

	mockService.EXPECT().
		UpdateAuthor(mock.Anything, authorID, requestBody).
		Return(nil).
		Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/authors/"+authorID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", authorID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.update(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthorHandler_Update_NotFound(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	authorID := "non-existent-id"
	requestBody := author.Request{
		FullName:  "Обновленный автор",
		Pseudonym: "Обновленный псевдоним",
		Specialty: "Обновленная специальность",
	}

	mockService.EXPECT().
		UpdateAuthor(mock.Anything, authorID, requestBody).
		Return(store.ErrorNotFound).
		Once()

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/authors/"+authorID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", authorID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.update(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAuthorHandler_Delete_Success(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	authorID := "123e4567-e89b-12d3-a456-426614174000"

	mockService.EXPECT().
		DeleteAuthor(mock.Anything, authorID).
		Return(nil).
		Once()

	req := httptest.NewRequest(http.MethodDelete, "/authors/"+authorID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", authorID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.delete(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuthorHandler_Delete_NotFound(t *testing.T) {
	mockService := mocks.NewMockAuthorService(t)
	handler := NewAuthorHandler(mockService)

	authorID := "non-existent-id"

	mockService.EXPECT().
		DeleteAuthor(mock.Anything, authorID).
		Return(store.ErrorNotFound).
		Once()

	req := httptest.NewRequest(http.MethodDelete, "/authors/"+authorID, nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", authorID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	handler.delete(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
