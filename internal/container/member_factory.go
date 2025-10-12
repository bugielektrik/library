package container

import (
	infraauth "library-service/internal/infrastructure/auth"
	memberdomain "library-service/internal/members/domain"
	memberauth "library-service/internal/members/service/auth"
	"library-service/internal/members/service/profile"
	"library-service/internal/members/service/subscription"
)

// ================================================================================
// Factory Functions - Member Domain
// ================================================================================

// newAuthUseCases creates all authentication-related use cases
func newAuthUseCases(
	memberRepo memberdomain.Repository,
	jwtService *infraauth.JWTService,
	passwordService *infraauth.PasswordService,
) AuthUseCases {
	// Create domain service
	memberService := memberdomain.NewService()

	return AuthUseCases{
		RegisterMember: memberauth.NewRegisterUseCase(memberRepo, passwordService, jwtService, memberService),
		LoginMember:    memberauth.NewLoginUseCase(memberRepo, passwordService, jwtService),
		RefreshToken:   memberauth.NewRefreshTokenUseCase(memberRepo, jwtService),
		ValidateToken:  memberauth.NewValidateTokenUseCase(memberRepo, jwtService),
	}
}

// newMemberUseCases creates all member-related use cases
func newMemberUseCases(memberRepo memberdomain.Repository) MemberUseCases {
	return MemberUseCases{
		ListMembers:      profile.NewListMembersUseCase(memberRepo),
		GetMemberProfile: profile.NewGetMemberProfileUseCase(memberRepo),
	}
}

// newSubscriptionUseCases creates subscription-related use cases
func newSubscriptionUseCases(memberRepo memberdomain.Repository) SubscriptionUseCases {
	// Create domain service
	memberService := memberdomain.NewService()

	return SubscriptionUseCases{
		SubscribeMember: subscription.NewSubscribeMemberUseCase(memberRepo, memberService),
	}
}
