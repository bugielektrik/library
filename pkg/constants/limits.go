package constants

// Pagination constants
const (
	// Default pagination values
	DefaultPageSize   = 20
	DefaultPageNumber = 1
	MaxPageSize       = 100
	MinPageSize       = 1

	// List limits
	DefaultListLimit = 50
	MaxListLimit     = 500
)

// Batch processing constants
const (
	// Payment batch processing
	DefaultPaymentBatchSize = 50
	MaxPaymentBatchSize     = 100

	// Callback retry batch processing
	DefaultRetryBatchSize = 50
	MaxRetryBatchSize     = 100

	// Expiry check batch size
	DefaultExpiryBatchSize = 100
	MaxExpiryBatchSize     = 500
)

// Retry constants
const (
	// Maximum retry attempts for various operations
	MaxPaymentRetries  = 3
	MaxCallbackRetries = 5
	MaxGatewayRetries  = 3
	MaxDatabaseRetries = 2

	// Retry delays (in seconds)
	InitialRetryDelay = 1
	MaxRetryDelay     = 60
)

// Validation limits
const (
	// String field lengths
	MinPasswordLength    = 8
	MaxPasswordLength    = 128
	MinNameLength        = 1
	MaxNameLength        = 255
	MaxEmailLength       = 320
	MaxDescriptionLength = 1000

	// ISBN constraints
	ISBN10Length = 10
	ISBN13Length = 13

	// Payment constraints
	MinPaymentAmount = 100      // 1.00 in smallest currency unit
	MaxPaymentAmount = 10000000 // 100,000.00 in smallest currency unit

	// Reservation constraints
	MaxActiveReservationsPerMember = 5
	MaxReservationDays             = 7
)

// Database constraints
const (
	// Query limits
	MaxInClauseItems  = 1000
	MaxBulkInsertRows = 1000

	// Connection pool settings
	DefaultMaxOpenConns    = 25
	DefaultMaxIdleConns    = 10
	DefaultConnMaxLifetime = 300 // seconds
)

// API rate limiting
const (
	// Requests per minute
	DefaultRateLimit         = 100
	AuthEndpointRateLimit    = 20
	PaymentEndpointRateLimit = 50

	// Burst size
	DefaultBurstSize = 10
	AuthBurstSize    = 5
)

// File upload limits
const (
	MaxFileUploadSize  = 10 * 1024 * 1024 // 10MB
	MaxImageUploadSize = 5 * 1024 * 1024  // 5MB
	MaxCSVUploadSize   = 2 * 1024 * 1024  // 2MB
)

// Cache sizes
const (
	DefaultCacheSize = 1000
	BookCacheSize    = 5000
	MemberCacheSize  = 2000
)
