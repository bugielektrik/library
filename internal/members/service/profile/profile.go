package profile

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/members/domain"
	"library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
)

// ================================================================================
// Get Member Profile Use Case
// ================================================================================

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

// ================================================================================
// List Members Use Case
// ================================================================================

// ListMembersRequest represents the input for listing members.
type ListMembersRequest struct {
	// Future: Add pagination, filters, sorting
}

// ListMembersResponse represents the output of listing members.
type ListMembersResponse struct {
	Members []domain.Member
	Total   int
}

// ListMembersUseCase handles listing all members.
type ListMembersUseCase struct {
	memberRepo domain.Repository
}

// NewListMembersUseCase creates a new instance of ListMembersUseCase.
func NewListMembersUseCase(memberRepo domain.Repository) *ListMembersUseCase {
	return &ListMembersUseCase{
		memberRepo: memberRepo,
	}
}

// Execute lists all members.
func (uc *ListMembersUseCase) Execute(ctx context.Context, req ListMembersRequest) (ListMembersResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "member", "list")

	members, err := uc.memberRepo.List(ctx)
	if err != nil {
		logger.Error("failed to list members", zap.Error(err))
		return ListMembersResponse{}, err
	}

	logger.Info("members listed successfully", zap.Int("count", len(members)))

	return ListMembersResponse{
		Members: members,
		Total:   len(members),
	}, nil
}
