package profile

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/members/domain"
	"library-service/pkg/logutil"
)

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
