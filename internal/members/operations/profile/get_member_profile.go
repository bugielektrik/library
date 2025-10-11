package profile

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/members/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// GetMemberProfileRequest represents the input for getting a member profile.
type GetMemberProfileRequest struct {
	MemberID string
}

// GetMemberProfileResponse represents the output of getting a member profile.
type GetMemberProfileResponse struct {
	Member domain.Member
}

// GetMemberProfileUseCase handles retrieving a member's profile.
type GetMemberProfileUseCase struct {
	memberRepo domain.Repository
}

// NewGetMemberProfileUseCase creates a new instance of GetMemberProfileUseCase.
func NewGetMemberProfileUseCase(memberRepo domain.Repository) *GetMemberProfileUseCase {
	return &GetMemberProfileUseCase{
		memberRepo: memberRepo,
	}
}

// Execute retrieves a member's profile.
func (uc *GetMemberProfileUseCase) Execute(ctx context.Context, req GetMemberProfileRequest) (GetMemberProfileResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "member", "get_profile")

	memberData, err := uc.memberRepo.Get(ctx, req.MemberID)
	if err != nil {
		logger.Error("failed to get member", zap.Error(err))
		return GetMemberProfileResponse{}, errors.ErrNotFound.WithDetails("member_id", req.MemberID)
	}

	logger.Info("member profile retrieved successfully")

	return GetMemberProfileResponse{
		Member: memberData,
	}, nil
}
