package v1

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"library-service/internal/adapters/http/dto"
	authuc "library-service/internal/usecase/auth"
	"library-service/pkg/errors"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	registerUseCase      *authuc.RegisterUseCase
	loginUseCase         *authuc.LoginUseCase
	refreshTokenUseCase  *authuc.RefreshTokenUseCase
	validateTokenUseCase *authuc.ValidateTokenUseCase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	registerUC *authuc.RegisterUseCase,
	loginUC *authuc.LoginUseCase,
	refreshUC *authuc.RefreshTokenUseCase,
	validateUC *authuc.ValidateTokenUseCase,
) *AuthHandler {
	return &AuthHandler{
		registerUseCase:      registerUC,
		loginUseCase:         loginUC,
		refreshTokenUseCase:  refreshUC,
		validateTokenUseCase: validateUC,
	}
}

// Routes returns the auth routes
func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.RefreshToken)
	r.Get("/me", h.GetCurrentMember) // Protected route - requires auth middleware

	return r
}

// Register handles member registration
// @Summary Register a new member
// @Description Create a new member account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authuc.RegisterRequest true "Registration details"
// @Success 201 {object} authuc.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "Email already exists"
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authuc.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Add details about the JSON decoding error
		h.respondError(w, errors.ErrInvalidInput.
			WithDetails("reason", "JSON decode failed").
			WithDetails("error", err.Error()))
		return
	}

	response, err := h.registerUseCase.Execute(r.Context(), req)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// Login handles member authentication
// @Summary Authenticate a member
// @Description Login with email and password to receive JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authuc.LoginRequest true "Login credentials"
// @Success 200 {object} authuc.LoginResponse
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authuc.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.ErrInvalidInput.Wrap(err))
		return
	}

	response, err := h.loginUseCase.Execute(r.Context(), req)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authuc.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} authuc.RefreshTokenResponse
// @Failure 401 {object} dto.ErrorResponse "Invalid or expired refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req authuc.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, errors.ErrInvalidInput.Wrap(err))
		return
	}

	response, err := h.refreshTokenUseCase.Execute(r.Context(), req)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// GetCurrentMember returns the current authenticated member's information
// @Summary Get current member
// @Description Get the authenticated member's information from JWT token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} authuc.ValidateTokenResponse
// @Failure 401 {object} dto.ErrorResponse "Invalid or missing token"
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentMember(w http.ResponseWriter, r *http.Request) {
	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.respondError(w, errors.ErrUnauthorized.WithDetails("reason", "missing authorization header"))
		return
	}

	// Check for Bearer prefix
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.respondError(w, errors.ErrUnauthorized.WithDetails("reason", "invalid authorization header format"))
		return
	}

	token := parts[1]

	// Validate token and get member info
	req := authuc.ValidateTokenRequest{
		AccessToken: token,
	}

	response, err := h.validateTokenUseCase.Execute(r.Context(), req)
	if err != nil {
		h.respondError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

// respondJSON sends a JSON response
func (h *AuthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Log error but response is already written
		_ = err
	}
}

// respondError sends an error response
func (h *AuthHandler) respondError(w http.ResponseWriter, err error) {
	status := errors.GetHTTPStatus(err)
	response := dto.FromError(err)
	h.respondJSON(w, status, response)
}