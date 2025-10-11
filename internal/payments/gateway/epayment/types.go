package epayment

// TokenResponse represents the OAuth token response from edomain.kz.
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"` // seconds
	Scope       string `json:"scope"`
}

// TransactionStatusResponse represents the response from transaction status check.
type TransactionStatusResponse struct {
	ResultCode    string             `json:"resultCode"`
	ResultMessage string             `json:"resultMessage"`
	Transaction   TransactionDetails `json:"transaction"`
}

// TransactionDetails contains detailed information about a transaction.
type TransactionDetails struct {
	ID           string `json:"id"`
	InvoiceID    string `json:"invoiceId"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Status       string `json:"status"`
	CardMask     string `json:"cardMask,omitempty"`
	ApprovalCode string `json:"approvalCode,omitempty"`
	Reference    string `json:"reference,omitempty"`
}

// RefundRequest represents a refund request.
type RefundRequest struct {
	TransactionID string
	Amount        *float64 // nil for full refund
	ExternalID    string   // optional tracking ID
}

// RefundResponse represents the response from refund API.
type RefundResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// CardPaymentRequest represents a request to charge a saved card.
type CardPaymentRequest struct {
	InvoiceID   string // Unique invoice identifier
	Amount      int64  // Amount in smallest currency unit (tenge)
	Currency    string // Currency code (e.g., "KZT")
	CardID      string // Saved card token
	Description string // Payment description
}

// CardPaymentResponse represents the response from card payment API.
type CardPaymentResponse struct {
	ID            string `json:"id"`
	TransactionID string `json:"transactionId"`
	Status        string `json:"status"`
	Reference     string `json:"reference,omitempty"`
	ApprovalCode  string `json:"approvalCode,omitempty"`
	ErrorCode     string `json:"errorCode,omitempty"`
	ErrorMessage  string `json:"errorMessage,omitempty"`
}
