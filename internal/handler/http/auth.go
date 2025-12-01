package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"library-service/internal/domain/user"
	"library-service/internal/service/auth"
	"library-service/internal/service/interfaces"
	"library-service/pkg/server/response"
)

type AuthHandler struct {
	authService interfaces.AuthService
}

func NewAuthHandler(s interfaces.AuthService) *AuthHandler {
	return &AuthHandler{authService: s}
}

func (h *AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/sign-up", h.signUp)
	r.Post("/sign-in", h.signIn)
	r.Post("/refresh", h.refresh)

	return r
}

// @Summary	sign up a new user
// @Tags		auth
// @Accept		json
// @Produce	json
// @Param		request	body		user.SignUpRequest	true	"body param"
// @Success	200		{object}	user.AuthResponse
// @Failure	400		{object}	response.Object
// @Failure	409		{object}	response.Object
// @Failure	500		{object}	response.Object
// @Router		/auth/sign-up [post]
func (h *AuthHandler) signUp(w http.ResponseWriter, r *http.Request) {
	req := user.SignUpRequest{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	res, err := h.authService.SignUp(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrUserAlreadyExists):
			response.Conflict(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}

	response.OK(w, r, res)
}

// @Summary	sign in existing user
// @Tags		auth
// @Accept		json
// @Produce	json
// @Param		request	body		user.SignInRequest	true	"body param"
// @Success	200		{object}	user.AuthResponse
// @Failure	400		{object}	response.Object
// @Failure	401		{object}	response.Object
// @Failure	500		{object}	response.Object
// @Router		/auth/sign-in [post]
func (h *AuthHandler) signIn(w http.ResponseWriter, r *http.Request) {
	req := user.SignInRequest{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	res, err := h.authService.SignIn(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			response.Unauthorized(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}

	response.OK(w, r, res)
}

// @Summary	refresh access token
// @Tags		auth
// @Accept		json
// @Produce	json
// @Param		request	body		user.RefreshRequest	true	"body param"
// @Success	200		{object}	user.AuthResponse
// @Failure	400		{object}	response.Object
// @Failure	401		{object}	response.Object
// @Failure	500		{object}	response.Object
// @Router		/auth/refresh [post]
func (h *AuthHandler) refresh(w http.ResponseWriter, r *http.Request) {
	req := user.RefreshRequest{}
	if err := render.Bind(r, &req); err != nil {
		response.BadRequest(w, r, err, req)
		return
	}

	res, err := h.authService.RefreshToken(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidToken):
			response.Unauthorized(w, r, err)
		default:
			response.InternalServerError(w, r, err, nil)
		}
		return
	}

	response.OK(w, r, res)
}
