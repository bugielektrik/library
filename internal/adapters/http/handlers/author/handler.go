package author

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/adapters/http/handlers"
	"library-service/internal/usecase"
	"library-service/internal/usecase/authorops"
	"library-service/pkg/logutil"
)

// AuthorHandler handles HTTP requests for authors
type AuthorHandler struct {
	handlers.BaseHandler
	useCases *usecase.Container
}

// NewAuthorHandler creates a new author handler
func NewAuthorHandler(
	useCases *usecase.Container,
) *AuthorHandler {
	return &AuthorHandler{
		useCases: useCases,
	}
}

// Routes returns the router for author endpoints
func (h *AuthorHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)

	return r
}

// @Summary List all authors
// @Description Retrieves a list of all authors in the system
// @Tags authors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.AuthorResponse "List of authors"
// @Failure 500 {object} dto.ErrorResponse "Internal server error"
// @Router /authors [get]
func (h *AuthorHandler) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "author_handler", "list")

	// Execute usecase
	result, err := h.useCases.Author.ListAuthors.Execute(ctx, authorops.ListAuthorsRequest{})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTOs
	authors := dto.FromAuthorEntities(result.Authors)

	logger.Info("authors listed", zap.Int("count", len(authors)))
	h.RespondJSON(w, http.StatusOK, authors)
}
