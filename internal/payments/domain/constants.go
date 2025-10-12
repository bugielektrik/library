package domain

import "time"

// Payment status transition rules
const (
	// Maximum time a payment can stay in pending/processing state
	PaymentExpirationTime = 30 * time.Minute

	// Time after which a completed payment cannot be refunded
	RefundWindowDuration = 30 * 24 * time.Hour // 30 days

	// Time after which a completed payment cannot be cancelled
	CancellationWindowDuration = 1 * time.Hour
)

// Payment provider constants
const (
	// Gateway request timeouts
	GatewayRequestTimeout = 30 * time.Second
	GatewayRetryTimeout   = 60 * time.Second

	// Gateway retry configuration
	MaxGatewayRetries = 3
	GatewayRetryDelay = 5 * time.Second

	// Token expiration buffer
	TokenExpirationBuffer = 5 * time.Minute
)

// Callback retry constants
const (
	// Maximum number of retry attempts for failed callbacks
	MaxCallbackRetries = 5

	// Delay between retry attempts (exponential backoff base)
	CallbackRetryBaseDelay = 1 * time.Minute

	// Maximum delay between retries
	CallbackRetryMaxDelay = 1 * time.Hour

	// Time after which a callback retry is considered failed permanently
	CallbackRetryTimeout = 24 * time.Hour
)

// Receipt generation constants
const (
	// Receipt number format prefix
	ReceiptNumberPrefix = "RCP"

	// Receipt validity period
	ReceiptValidityPeriod = 365 * 24 * time.Hour // 1 year
)

// Saved card constants
const (
	// Maximum saved cards per member
	MaxSavedCardsPerMember = 10

	// Card expiration warning period
	CardExpirationWarningPeriod = 30 * 24 * time.Hour // 30 days
)

// Payment amount constraints
const (
	// Minimum payment amount in smallest currency unit (e.g., cents)
	MinPaymentAmount = 100 // 1.00

	// Maximum payment amount in smallest currency unit
	MaxPaymentAmount = 10000000 // 100,000.00

	// Maximum refund percentage of original payment
	MaxRefundPercentage = 100
)

// Subscription payment constants
const (
	// Subscription payment types
	SubscriptionMonthly   = "monthly"
	SubscriptionQuarterly = "quarterly"
	SubscriptionAnnual    = "annual"

	// Subscription durations
	MonthlySubscriptionDuration   = 30 * 24 * time.Hour
	QuarterlySubscriptionDuration = 90 * 24 * time.Hour
	AnnualSubscriptionDuration    = 365 * 24 * time.Hour

	// Subscription prices in cents
	MonthlySubscriptionPrice   = 1000  // 10.00
	QuarterlySubscriptionPrice = 2700  // 27.00 (10% discount)
	AnnualSubscriptionPrice    = 10000 // 100.00 (17% discount)
)

// Fine payment constants
const (
	// Fine calculation rates
	DailyFineRate = 50   // 0.50 per day in cents
	MaxFineAmount = 5000 // 50.00 maximum fine

	// Grace period before fines apply
	FineGracePeriod = 24 * time.Hour
)

// Currency constants
const (
	DefaultCurrency = "KZT"

	// Supported currencies
	CurrencyKZT = "KZT"
	CurrencyUSD = "USD"
	CurrencyEUR = "EUR"
	CurrencyRUB = "RUB"
)

// Payment method constants
const (
	// Payment method types
	PaymentMethodCard         = "card"
	PaymentMethodSavedCard    = "saved_card"
	PaymentMethodWallet       = "wallet"
	PaymentMethodBankTransfer = "bank_transfer"
)

// Transaction status codes from provider
const (
	GatewayStatusSuccess    = "Success"
	GatewayStatusProcessing = "Processing"
	GatewayStatusFailed     = "Failed"
	GatewayStatusCancelled  = "Cancelled"
	GatewayStatusRefunded   = "Refunded"
)

// Error codes for payment operations
const (
	ErrCodeInsufficientFunds = "INSUFFICIENT_FUNDS"
	ErrCodeCardExpired       = "CARD_EXPIRED"
	ErrCodeCardDeclined      = "CARD_DECLINED"
	ErrCodeInvalidAmount     = "INVALID_AMOUNT"
	ErrCodeDuplicatePayment  = "DUPLICATE_PAYMENT"
	ErrCodeGatewayTimeout    = "GATEWAY_TIMEOUT"
	ErrCodeGatewayError      = "GATEWAY_ERROR"
)
