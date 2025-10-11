// Package subscription implements use cases for subscription management operations.
//
// This package handles member subscription workflows including subscription
// creation, renewal, cancellation, and upgrade/downgrade between tiers.
// Subscriptions enable members to access library services and borrow books.
//
// Use cases implemented:
//   - SubscribeMemberUseCase: Creates or renews member subscription
//   - UpgradeSubscriptionUseCase: Upgrades member to higher tier
//   - CancelSubscriptionUseCase: Cancels active subscription
//   - GetSubscriptionStatusUseCase: Retrieves current subscription info
//
// Dependencies:
//   - member.Repository: For updating member subscription status
//   - member.Service: For subscription pricing and business rules
//   - payment.Gateway: For processing subscription payments
//
// Example usage:
//
//	subscribeUC := subscription.NewSubscribeMemberUseCase(memberRepo, memberService, paymentGateway)
//	response, err := subscribeUC.Execute(ctx, subscription.SubscribeRequest{
//	    MemberID: "member-uuid",
//	    Tier:     "premium",
//	    Duration: "monthly",
//	})
//	// response contains: SubscriptionID, ExpiresAt, PaymentURL
//
// Subscription tiers:
//   - Basic ($4.99/month): 3 concurrent loans, standard queue priority
//   - Premium ($9.99/month): 10 concurrent loans, priority queue, no late fees
//
// Subscription duration:
//   - Monthly: Billed every 30 days
//   - Annual: Billed yearly with 2 months discount
//
// Business rules:
//   - Payment must be successful before subscription activation
//   - Expired subscriptions result in loan privileges suspended
//   - Downgrade takes effect at end of current billing period
//   - Upgrade prorated based on remaining days in current period
//   - Auto-renewal enabled by default, can be disabled
//
// Payment integration:
//   - Initial subscription requires payment
//   - Renewal attempts 3 days before expiry
//   - Failed renewals trigger grace period (7 days)
//   - Email notifications sent before expiry
//
// Architecture:
//   - Package organized in bounded context structure
//   - Pricing logic delegated to member.Service
//   - Payment processing via payment gateway adapter
//   - Subscription state stored in member entity
package subscription
