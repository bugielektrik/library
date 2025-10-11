package member

import (
	"testing"
	"time"

	"library-service/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestService_ValidateSubscriptionType(t *testing.T) {
	service := NewService()

	tests := []struct {
		name      string
		subType   string
		wantError bool
	}{
		{
			name:      "valid basic subscription",
			subType:   "basic",
			wantError: false,
		},
		{
			name:      "valid premium subscription",
			subType:   "premium",
			wantError: false,
		},
		{
			name:      "valid annual subscription",
			subType:   "annual",
			wantError: false,
		},
		{
			name:      "invalid subscription type",
			subType:   "platinum",
			wantError: true,
		},
		{
			name:      "empty subscription type",
			subType:   "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSubscriptionType(tt.subType)

			if tt.wantError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errors.ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_ValidateSubscriptionDuration(t *testing.T) {
	service := NewService()

	tests := []struct {
		name      string
		months    int
		wantError bool
	}{
		{
			name:      "valid 1 month",
			months:    1,
			wantError: false,
		},
		{
			name:      "valid 12 months",
			months:    12,
			wantError: false,
		},
		{
			name:      "valid 24 months (max)",
			months:    24,
			wantError: false,
		},
		{
			name:      "invalid 0 months",
			months:    0,
			wantError: true,
		},
		{
			name:      "invalid negative months",
			months:    -1,
			wantError: true,
		},
		{
			name:      "invalid too many months",
			months:    25,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateSubscriptionDuration(tt.months)

			if tt.wantError {
				assert.Error(t, err)
				assert.ErrorIs(t, err, errors.ErrInvalidInput)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CalculateSubscriptionPrice(t *testing.T) {
	service := NewService()

	tests := []struct {
		name          string
		subType       string
		months        int
		expectedPrice float64
		wantError     bool
	}{
		{
			name:          "basic 1 month - no discount",
			subType:       "basic",
			months:        1,
			expectedPrice: 9.99,
			wantError:     false,
		},
		{
			name:          "premium 1 month - no discount",
			subType:       "premium",
			months:        1,
			expectedPrice: 19.99,
			wantError:     false,
		},
		{
			name:          "basic 6 months - 10% discount",
			subType:       "basic",
			months:        6,
			expectedPrice: 9.99 * 6 * 0.90,
			wantError:     false,
		},
		{
			name:          "premium 12 months - 20% discount",
			subType:       "premium",
			months:        12,
			expectedPrice: 19.99 * 12 * 0.80,
			wantError:     false,
		},
		{
			name:          "annual 12 months - 20% discount",
			subType:       "annual",
			months:        12,
			expectedPrice: 14.99 * 12 * 0.80,
			wantError:     false,
		},
		{
			name:          "invalid subscription type",
			subType:       "invalid",
			months:        1,
			expectedPrice: 0,
			wantError:     true,
		},
		{
			name:          "invalid duration",
			subType:       "basic",
			months:        0,
			expectedPrice: 0,
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := service.CalculateSubscriptionPrice(tt.subType, tt.months)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expectedPrice, price, 0.01)
			}
		})
	}
}

func TestService_CalculateExpirationDate(t *testing.T) {
	service := NewService()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		months       int
		expectedDate time.Time
	}{
		{
			name:         "1 month subscription",
			months:       1,
			expectedDate: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "6 months subscription",
			months:       6,
			expectedDate: time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name:         "12 months subscription",
			months:       12,
			expectedDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateExpirationDate(startDate, tt.months)
			assert.Equal(t, tt.expectedDate, result)
		})
	}
}

func TestService_IsSubscriptionActive(t *testing.T) {
	service := NewService()

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	tomorrow := now.AddDate(0, 0, 1)

	tests := []struct {
		name         string
		subscribedAt time.Time
		expiresAt    time.Time
		expected     bool
	}{
		{
			name:         "active subscription",
			subscribedAt: yesterday,
			expiresAt:    tomorrow,
			expected:     true,
		},
		{
			name:         "expired subscription",
			subscribedAt: yesterday,
			expiresAt:    yesterday.Add(1 * time.Hour),
			expected:     false,
		},
		{
			name:         "future subscription",
			subscribedAt: tomorrow,
			expiresAt:    tomorrow.AddDate(0, 1, 0),
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsSubscriptionActive(tt.subscribedAt, tt.expiresAt)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestService_CanUpgradeSubscription(t *testing.T) {
	service := NewService()

	tests := []struct {
		name        string
		currentType string
		targetType  string
		wantError   bool
	}{
		{
			name:        "upgrade from basic to premium",
			currentType: "basic",
			targetType:  "premium",
			wantError:   false,
		},
		{
			name:        "upgrade from basic to annual",
			currentType: "basic",
			targetType:  "annual",
			wantError:   false,
		},
		{
			name:        "same tier - premium to annual",
			currentType: "premium",
			targetType:  "annual",
			wantError:   false,
		},
		{
			name:        "downgrade from premium to basic",
			currentType: "premium",
			targetType:  "basic",
			wantError:   true,
		},
		{
			name:        "invalid current type",
			currentType: "invalid",
			targetType:  "premium",
			wantError:   true,
		},
		{
			name:        "invalid target type",
			currentType: "basic",
			targetType:  "invalid",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CanUpgradeSubscription(tt.currentType, tt.targetType)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_Validate(t *testing.T) {
	service := NewService()

	validName := "John Doe"
	emptyString := ""

	tests := []struct {
		name      string
		member    Member
		wantError bool
		errorType *errors.Error
	}{
		{
			name: "valid member",
			member: Member{
				FullName: &validName,
			},
			wantError: false,
		},
		{
			name: "missing full name",
			member: Member{
				FullName: nil,
			},
			wantError: true,
			errorType: errors.ErrInvalidMemberData,
		},
		{
			name: "empty full name",
			member: Member{
				FullName: &emptyString,
			},
			wantError: true,
			errorType: errors.ErrInvalidMemberData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.Validate(tt.member)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorType != nil {
					assert.ErrorIs(t, err, tt.errorType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_CalculateGracePeriod(t *testing.T) {
	service := NewService()

	tests := []struct {
		name     string
		subType  string
		expected time.Duration
	}{
		{
			name:     "premium gets 7 days",
			subType:  "premium",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "annual gets 7 days",
			subType:  "annual",
			expected: 7 * 24 * time.Hour,
		},
		{
			name:     "basic gets 3 days",
			subType:  "basic",
			expected: 3 * 24 * time.Hour,
		},
		{
			name:     "unknown type gets 0",
			subType:  "unknown",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.CalculateGracePeriod(tt.subType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestService_IsWithinGracePeriod(t *testing.T) {
	service := NewService()

	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		subType   string
		expected  bool
	}{
		{
			name:      "still active - not in grace period",
			expiresAt: now.Add(24 * time.Hour),
			subType:   "premium",
			expected:  false,
		},
		{
			name:      "expired 1 day ago - within premium grace period (7 days)",
			expiresAt: now.Add(-24 * time.Hour),
			subType:   "premium",
			expected:  true,
		},
		{
			name:      "expired 8 days ago - beyond premium grace period",
			expiresAt: now.Add(-8 * 24 * time.Hour),
			subType:   "premium",
			expected:  false,
		},
		{
			name:      "expired 2 days ago - within basic grace period (3 days)",
			expiresAt: now.Add(-2 * 24 * time.Hour),
			subType:   "basic",
			expected:  true,
		},
		{
			name:      "expired 4 days ago - beyond basic grace period",
			expiresAt: now.Add(-4 * 24 * time.Hour),
			subType:   "basic",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.IsWithinGracePeriod(tt.expiresAt, tt.subType)
			assert.Equal(t, tt.expected, result)
		})
	}
}
