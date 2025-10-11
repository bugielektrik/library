package usecase

import (
	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/internal/usecase/authops"
	"library-service/internal/usecase/memberops"
	"library-service/internal/usecase/subops"
)

// AuthUseCases contains all authentication-related use cases
type AuthUseCases struct {
	RegisterMember *authops.RegisterUseCase
	LoginMember    *authops.LoginUseCase
	RefreshToken   *authops.RefreshTokenUseCase
	ValidateToken  *authops.ValidateTokenUseCase
}

// MemberUseCases contains all member-related use cases
type MemberUseCases struct {
	ListMembers      *memberops.ListMembersUseCase
	GetMemberProfile *memberops.GetMemberProfileUseCase
}

// SubscriptionUseCases contains subscription-related use cases
type SubscriptionUseCases struct {
	SubscribeMember *subops.SubscribeMemberUseCase
}

// newAuthUseCases creates all authentication-related use cases
func newAuthUseCases(
	memberRepo member.Repository,
	jwtService *auth.JWTService,
	passwordService *auth.PasswordService,
) AuthUseCases {
	// Create domain service
	memberService := member.NewService()

	return AuthUseCases{
		RegisterMember: authops.NewRegisterUseCase(memberRepo, passwordService, jwtService, memberService),
		LoginMember:    authops.NewLoginUseCase(memberRepo, passwordService, jwtService),
		RefreshToken:   authops.NewRefreshTokenUseCase(memberRepo, jwtService),
		ValidateToken:  authops.NewValidateTokenUseCase(memberRepo, jwtService),
	}
}

// newMemberUseCases creates all member-related use cases
func newMemberUseCases(memberRepo member.Repository) MemberUseCases {
	return MemberUseCases{
		ListMembers:      memberops.NewListMembersUseCase(memberRepo),
		GetMemberProfile: memberops.NewGetMemberProfileUseCase(memberRepo),
	}
}

// newSubscriptionUseCases creates subscription-related use cases
func newSubscriptionUseCases(memberRepo member.Repository) SubscriptionUseCases {
	// Create domain service
	memberService := member.NewService()

	return SubscriptionUseCases{
		SubscribeMember: subops.NewSubscribeMemberUseCase(memberRepo, memberService),
	}
}
