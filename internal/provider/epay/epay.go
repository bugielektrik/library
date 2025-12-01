package epay

import "net/http"

// Provider defines the interface for payment provider operations
type Provider interface {
	// Payment operations
	PayByTemplate(w http.ResponseWriter, request Request) error
	PayByCard(request Request) (*ResponseCardID, error)

	// Transaction operations
	CheckStatus(invoiceID, terminalID string) (StatusResponse, error)
	Charge(transactionID, amount string) error
	Cancel(transactionID string) error

	// Token management
	InitTokenRefresher() error
}

// Token represents OAuth2 access token response
type Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

// Credentials holds authentication credentials for the payment provider
type Credentials struct {
	Username string
	Password string
	Endpoint string
	OAuth    string
	JS       string

	AccessToken string
	ExpiresIn   int64
}
