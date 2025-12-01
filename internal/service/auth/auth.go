package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"library-service/config"
	"library-service/internal/domain/user"
	"library-service/pkg/log"
	"library-service/pkg/store"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	userRepository user.Repository
	jwtConfig      config.JWTConfig
}

func NewAuthService(r user.Repository, cfg config.JWTConfig) *AuthService {
	return &AuthService{
		userRepository: r,
		jwtConfig:      cfg,
	}
}

func (s *AuthService) SignUp(ctx context.Context, req user.SignUpRequest) (user.AuthResponse, error) {
	logger := log.FromContext(ctx).Named("sign_up").With(zap.String("email", req.Email))

	existingUser, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, store.ErrorNotFound) {
		logger.Error("failed to check existing user", zap.Error(err))
		return user.AuthResponse{}, err
	}

	if existingUser.ID != "" {
		logger.Warn("user already exists")
		return user.AuthResponse{}, ErrUserAlreadyExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return user.AuthResponse{}, err
	}

	newUser := user.New(req, string(hashedPassword))
	newUser.ID = uuid.New().String()

	id, err := s.userRepository.Create(ctx, newUser)
	if err != nil {
		logger.Error("failed to create user", zap.Error(err))
		return user.AuthResponse{}, err
	}
	newUser.ID = id

	accessToken, err := s.generateAccessToken(newUser.ID)
	if err != nil {
		logger.Error("failed to generate access token", zap.Error(err))
		return user.AuthResponse{}, err
	}

	refreshToken, err := s.generateRefreshToken(newUser.ID)
	if err != nil {
		logger.Error("failed to generate refresh token", zap.Error(err))
		return user.AuthResponse{}, err
	}

	return user.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ParseFromEntity(newUser),
	}, nil
}

func (s *AuthService) SignIn(ctx context.Context, req user.SignInRequest) (user.AuthResponse, error) {
	logger := log.FromContext(ctx).Named("sign_in").With(zap.String("email", req.Email))

	userEntity, err := s.userRepository.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("user not found")
			return user.AuthResponse{}, ErrInvalidCredentials
		}
		logger.Error("failed to get user", zap.Error(err))
		return user.AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(userEntity.PasswordHash), []byte(req.Password)); err != nil {
		logger.Warn("invalid password")
		return user.AuthResponse{}, ErrInvalidCredentials
	}

	accessToken, err := s.generateAccessToken(userEntity.ID)
	if err != nil {
		logger.Error("failed to generate access token", zap.Error(err))
		return user.AuthResponse{}, err
	}

	refreshToken, err := s.generateRefreshToken(userEntity.ID)
	if err != nil {
		logger.Error("failed to generate refresh token", zap.Error(err))
		return user.AuthResponse{}, err
	}

	return user.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ParseFromEntity(userEntity),
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, req user.RefreshRequest) (user.AuthResponse, error) {
	logger := log.FromContext(ctx).Named("refresh_token")

	claims, err := s.validateRefreshToken(req.RefreshToken)
	if err != nil {
		logger.Warn("invalid refresh token", zap.Error(err))
		return user.AuthResponse{}, ErrInvalidToken
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		logger.Error("invalid user_id in token")
		return user.AuthResponse{}, ErrInvalidToken
	}

	userEntity, err := s.userRepository.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("user not found")
			return user.AuthResponse{}, ErrInvalidToken
		}
		logger.Error("failed to get user", zap.Error(err))
		return user.AuthResponse{}, err
	}

	accessToken, err := s.generateAccessToken(userEntity.ID)
	if err != nil {
		logger.Error("failed to generate access token", zap.Error(err))
		return user.AuthResponse{}, err
	}

	refreshToken, err := s.generateRefreshToken(userEntity.ID)
	if err != nil {
		logger.Error("failed to generate refresh token", zap.Error(err))
		return user.AuthResponse{}, err
	}

	return user.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user.ParseFromEntity(userEntity),
	}, nil
}

func (s *AuthService) generateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.jwtConfig.AccessTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.AccessSecret))
}

func (s *AuthService) generateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(s.jwtConfig.RefreshTokenTTL).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.RefreshSecret))
}

func (s *AuthService) validateRefreshToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.jwtConfig.RefreshSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
