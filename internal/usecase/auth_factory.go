package usecase

import (
	infraauth "library-service/internal/infrastructure/auth"
	"library-service/internal/members/domain"
	memberauth "library-service/internal/members/operations/auth"
	"library-service/internal/members/operations/profile"
	"library-service/internal/members/operations/subscription"
)

// AuthUseCases contains all authentication-related use cases
type AuthUseCases struct {
	RegisterMember *memberauth.RegisterUseCase
	LoginMember    *memberauth.LoginUseCase
	RefreshToken   *memberauth.RefreshTokenUseCase
	ValidateToken  *memberauth.ValidateTokenUseCase
}

// MemberUseCases contains all member-related use cases
type MemberUseCases struct {
	ListMembers      *profile.ListMembersUseCase
	GetMemberProfile *profile.GetMemberProfileUseCase
}

// SubscriptionUseCases contains subscription-related use cases
type SubscriptionUseCases struct {
	SubscribeMember *subscription.SubscribeMemberUseCase
}

// newAuthUseCases creates all authentication-related use cases
func newAuthUseCases(
	memberRepo domain.Repository,
	jwtService *infraauth.JWTService,
	passwordService *infraauth.PasswordService,
) AuthUseCases {
	// Create domain service
	memberService := domain.NewService()

	return AuthUseCases{
		RegisterMember: memberauth.NewRegisterUseCase(memberRepo, passwordService, jwtService, memberService),
		LoginMember:    memberauth.NewLoginUseCase(memberRepo, passwordService, jwtService),
		RefreshToken:   memberauth.NewRefreshTokenUseCase(memberRepo, jwtService),
		ValidateToken:  memberauth.NewValidateTokenUseCase(memberRepo, jwtService),
	}
}

// newMemberUseCases creates all member-related use cases
func newMemberUseCases(memberRepo domain.Repository) MemberUseCases {
	return MemberUseCases{
		ListMembers:      profile.NewListMembersUseCase(memberRepo),
		GetMemberProfile: profile.NewGetMemberProfileUseCase(memberRepo),
	}
}

// newSubscriptionUseCases creates subscription-related use cases
func newSubscriptionUseCases(memberRepo domain.Repository) SubscriptionUseCases {
	// Create domain service
	memberService := domain.NewService()

	return SubscriptionUseCases{
		SubscribeMember: subscription.NewSubscribeMemberUseCase(memberRepo, memberService),
	}
}
