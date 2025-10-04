package subscription

import (
	"context"
	"time"

	"go.uber.org/zap"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/log"
	"library-service/internal/infrastructure/store"
	"library-service/pkg/errors"
)

// SubscribeMemberRequest represents the input for subscribing a member
type SubscribeMemberRequest struct {
	MemberID         string
	SubscriptionType string // e.g., "basic", "premium", "annual"
	DurationMonths   int
}

// SubscribeMemberResponse represents the output of subscribing a member
type SubscribeMemberResponse struct {
	MemberID         string
	SubscribedAt     time.Time
	ExpiresAt        time.Time
	Status           string
	SubscriptionType string
}

// SubscribeMemberUseCase handles the complex workflow of subscribing a member
// This involves:
// 1. Validating member exists
// 2. Checking for existing active subscriptions
// 3. Creating the subscription
// 4. Sending welcome email (future)
// 5. Publishing subscription event (future)
type SubscribeMemberUseCase struct {
	memberRepo    member.Repository
	memberService *member.Service
	// Future: emailService EmailService
	// Future: eventPublisher EventPublisher
}

// NewSubscribeMemberUseCase creates a new instance of SubscribeMemberUseCase
func NewSubscribeMemberUseCase(memberRepo member.Repository, memberService *member.Service) *SubscribeMemberUseCase {
	return &SubscribeMemberUseCase{
		memberRepo:    memberRepo,
		memberService: memberService,
	}
}

// Execute subscribes a member to the library service
func (uc *SubscribeMemberUseCase) Execute(ctx context.Context, req SubscribeMemberRequest) (SubscribeMemberResponse, error) {
	logger := log.FromContext(ctx).Named("subscribe_member_usecase").With(
		zap.String("member_id", req.MemberID),
		zap.String("subscription_type", req.SubscriptionType),
	)

	// Step 1: Validate request using domain service
	if err := uc.memberService.ValidateSubscriptionType(req.SubscriptionType); err != nil {
		logger.Warn("invalid subscription type", zap.Error(err))
		return SubscribeMemberResponse{}, err
	}

	if err := uc.memberService.ValidateSubscriptionDuration(req.DurationMonths); err != nil {
		logger.Warn("invalid subscription duration", zap.Error(err))
		return SubscribeMemberResponse{}, err
	}

	if req.MemberID == "" {
		err := errors.ErrInvalidInput.WithDetails("field", "member_id")
		logger.Warn("validation failed", zap.Error(err))
		return SubscribeMemberResponse{}, err
	}

	// Step 2: Verify member exists
	memberEntity, err := uc.memberRepo.Get(ctx, req.MemberID)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("member not found")
			return SubscribeMemberResponse{}, errors.ErrMemberNotFound.WithDetails("id", req.MemberID)
		}
		logger.Error("failed to get member", zap.Error(err))
		return SubscribeMemberResponse{}, errors.ErrDatabase.Wrap(err)
	}

	// Step 3: Check for active subscription (business rule)
	// In a real system, this would check a subscriptions table
	// For now, we'll simulate this check
	if uc.hasActiveSubscription(memberEntity) {
		logger.Warn("member already has active subscription")
		return SubscribeMemberResponse{}, errors.ErrSubscriptionActive.WithDetails("member_id", req.MemberID)
	}

	// Step 4: Create subscription using domain service
	now := time.Now()
	expiresAt := uc.memberService.CalculateExpirationDate(now, req.DurationMonths)

	// In a real system, you would:
	// - Create subscription record in store
	// - Process payment
	// - Grant access to library features
	// - Send confirmation email
	// - Publish domain event

	logger.Info("member subscribed successfully",
		zap.Time("subscribed_at", now),
		zap.Time("expires_at", expiresAt),
	)

	// Step 5: Return response
	return SubscribeMemberResponse{
		MemberID:         req.MemberID,
		SubscribedAt:     now,
		ExpiresAt:        expiresAt,
		Status:           "active",
		SubscriptionType: req.SubscriptionType,
	}, nil
}

// hasActiveSubscription checks if member has an active subscription
// In a real system, this would query a subscriptions table
func (uc *SubscribeMemberUseCase) hasActiveSubscription(m member.Member) bool {
	// Placeholder: In production, check against subscriptions repository
	// For now, always return false to allow subscriptions
	return false
}
