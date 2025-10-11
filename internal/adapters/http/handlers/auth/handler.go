package auth

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"library-service/internal/adapters/http/handlers"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/usecase"
	"library-service/internal/usecase/authops"
	"library-service/pkg/errors"
	"library-service/pkg/httputil"
	"library-service/pkg/logutil"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	handlers.BaseHandler
	useCases  *usecase.Container
	validator *middleware.Validator
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(
	useCases *usecase.Container,
	validator *middleware.Validator,
) *AuthHandler {
	return &AuthHandler{
		useCases:  useCases,
		validator: validator,
	}
}

// Routes returns the auth routes with selective authentication
func (h *AuthHandler) Routes(authMiddleware interface {
	Authenticate(http.Handler) http.Handler
}) chi.Router {
	r := chi.NewRouter()

	// Public routes
	r.Post("/register", h.register)
	r.Post("/login", h.login)
	r.Post("/refresh", h.refreshToken)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware.Authenticate)
		r.Get("/me", h.getCurrentMember)
	})

	return r
}

// register handles member registration
// @Summary Register a new member
// @Description Create a new member account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authops.RegisterRequest true "Registration details"
// @Success 201 {object} authops.RegisterResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse "Email already exists"
// @Router /auth/register [post]
func (h *AuthHandler) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "auth_handler", "register")

	var req authops.RegisterRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	if !h.validator.ValidateStruct(w, req) {
		return
	}

	response, err := h.useCases.Auth.RegisterMember.Execute(ctx, req)
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("member registered", zap.String("member_id", response.Member.ID))
	h.RespondJSON(w, http.StatusCreated, response)
}

// login handles member authentication
// @Summary Authenticate a member
// @Description Login with email and password to receive JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authops.LoginRequest true "Login credentials"
// @Success 200 {object} authops.LoginResponse
// @Failure 401 {object} dto.ErrorResponse "Invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "auth_handler", "login")

	var req authops.LoginRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	if !h.validator.ValidateStruct(w, req) {
		return
	}

	response, err := h.useCases.Auth.LoginMember.Execute(ctx, req)
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("member logged in", zap.String("member_id", response.Member.ID))
	h.RespondJSON(w, http.StatusOK, response)
}

// refreshToken handles token refresh
// @Summary Refresh access token
// @Description Exchange a valid refresh token for a new access token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body authops.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} authops.RefreshTokenResponse
// @Failure 401 {object} dto.ErrorResponse "Invalid or expired refresh token"
// @Router /auth/refresh [post]
func (h *AuthHandler) refreshToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "auth_handler", "refresh_token")

	var req authops.RefreshTokenRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	if !h.validator.ValidateStruct(w, req) {
		return
	}

	response, err := h.useCases.Auth.RefreshToken.Execute(ctx, req)
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("token refreshed successfully")
	h.RespondJSON(w, http.StatusOK, response)
}

// getCurrentMember returns the current authenticated member's information
// @Summary Get current member
// @Description Get the authenticated member's information from JWT token
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} authops.ValidateTokenResponse
// @Failure 401 {object} dto.ErrorResponse "Invalid or missing token"
// @Router /auth/me [get]
func (h *AuthHandler) getCurrentMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "auth_handler", "get_current_member")

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		h.RespondError(w, r, errors.ErrUnauthorized.WithDetails("reason", "missing authorization header"))
		return
	}

	// Check for Bearer prefix
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		h.RespondError(w, r, errors.ErrUnauthorized.WithDetails("reason", "invalid authorization header format"))
		return
	}

	token := parts[1]

	// Validate token and get member info
	req := authops.ValidateTokenRequest{
		AccessToken: token,
	}

	response, err := h.useCases.Auth.ValidateToken.Execute(ctx, req)
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("current member retrieved", zap.String("member_id", response.Member.ID))
	h.RespondJSON(w, http.StatusOK, response)
}
