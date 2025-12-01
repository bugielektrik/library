package user

import (
	"errors"
	"net/http"
	"regexp"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type SignUpRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
}

func (s *SignUpRequest) Bind(r *http.Request) error {
	if s.Email == "" {
		return errors.New("email: cannot be blank")
	}

	if !emailRegex.MatchString(s.Email) {
		return errors.New("email: invalid format")
	}

	if s.Password == "" {
		return errors.New("password: cannot be blank")
	}

	if len(s.Password) < 6 {
		return errors.New("password: must be at least 6 characters")
	}

	if s.FullName == "" {
		return errors.New("fullName: cannot be blank")
	}

	return nil
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *SignInRequest) Bind(r *http.Request) error {
	if s.Email == "" {
		return errors.New("email: cannot be blank")
	}

	if s.Password == "" {
		return errors.New("password: cannot be blank")
	}

	return nil
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func (s *RefreshRequest) Bind(r *http.Request) error {
	if s.RefreshToken == "" {
		return errors.New("refreshToken: cannot be blank")
	}

	return nil
}

type AuthResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         UserResponse
}

type UserResponse struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"fullName"`
}

func ParseFromEntity(data Entity) UserResponse {
	fullName := ""
	if data.FullName != nil {
		fullName = *data.FullName
	}

	return UserResponse{
		ID:       data.ID,
		Email:    data.Email,
		FullName: fullName,
	}
}
