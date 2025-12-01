package interfaces

import (
	"context"
	"library-service/internal/domain/user"
)

type AuthService interface {
	SignUp(ctx context.Context, req user.SignUpRequest) (user.AuthResponse, error)
	SignIn(ctx context.Context, req user.SignInRequest) (user.AuthResponse, error)
	RefreshToken(ctx context.Context, req user.RefreshRequest) (user.AuthResponse, error)
}
