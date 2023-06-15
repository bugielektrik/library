package auth

import (
	"errors"
	"net/http"

	"github.com/go-chi/oauth"
)

// ValidateUser validates username and password returning an error if the user credentials are wrong
func (*Service) ValidateUser(username, password, scope string, r *http.Request) error {
	if username == "user01" && password == "12345" {
		return nil
	}

	return errors.New("wrong user")
}

// ValidateClient validates clientID and secret returning an error if the client credentials are wrong
func (*Service) ValidateClient(clientID, clientSecret, scope string, r *http.Request) error {
	if clientID == "abcdef" && clientSecret == "12345" {
		return nil
	}

	return errors.New("wrong client")
}

// ValidateCode validates token ID
func (*Service) ValidateCode(clientID, clientSecret, code, redirectURI string, r *http.Request) (string, error) {
	return "", nil
}

// AddClaims provides additional claims to the token
func (*Service) AddClaims(tokenType oauth.TokenType, credential, tokenID, scope string, r *http.Request) (map[string]string, error) {
	claims := make(map[string]string)
	claims["id"] = "1001"
	claims["data"] = `{"order_date":"2016-12-14","order_id":"9999"}`
	return claims, nil
}

// AddProperties provides additional information to the token response
func (*Service) AddProperties(tokenType oauth.TokenType, credential, tokenID, scope string, r *http.Request) (map[string]string, error) {
	props := make(map[string]string)
	props["name"] = "Gopher"
	return props, nil
}

// ValidateTokenID validates token ID
func (*Service) ValidateTokenID(tokenType oauth.TokenType, credential, tokenID, refreshTokenID string) error {
	return nil
}

// StoreTokenID saves the token id generated for the user
func (*Service) StoreTokenID(tokenType oauth.TokenType, credential, tokenID, refreshTokenID string) error {
	return nil
}
