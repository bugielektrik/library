// Package memberops implements use cases for member management operations.
//
// This package orchestrates member-related workflows including member profile
// management, subscription handling, and member information retrieval. Members
// are registered users who can borrow books and access library services.
//
// Use cases implemented:
//   - GetMemberUseCase: Retrieves member details by ID
//   - UpdateMemberUseCase: Updates member profile information
//   - ListMembersUseCase: Returns all members (admin operation)
//   - SubscribeMemberUseCase: Handles member subscription workflows
//
// Dependencies:
//   - member.Repository: For member persistence
//   - member.Service: For subscription business rules and pricing
//
// Example usage:
//
//	getMemberUC := memberops.NewGetMemberUseCase(repo)
//	response, err := getMemberUC.Execute(ctx, memberops.GetMemberRequest{
//	    ID: "member-uuid",
//	})
//
// Subscription handling:
//   - Premium tier: $9.99/month, allows 10 concurrent loans
//   - Basic tier: $4.99/month, allows 3 concurrent loans
//   - Subscription expiry tracked and validated
//   - Auto-renewal configurable
//
// Business rules:
//   - Email must be unique across all members
//   - Active subscription required for book borrowing
//   - Member profile updates logged for audit trail
//
// Architecture:
//   - Package name uses "ops" suffix to avoid conflict with domain member package
//   - Subscription logic delegated to member.Service in domain layer
//   - Member creation handled by authops.RegisterMemberUseCase
package memberops
