package container

import (
	memberdomain "library-service/internal/members/domain"
	paymentdomain "library-service/internal/payments/domain"
	paymentservice "library-service/internal/payments/service/payment"
	receiptservice "library-service/internal/payments/service/receipt"
	savedcardservice "library-service/internal/payments/service/savedcard"
)

// ================================================================================
// Factory Functions - Payment Domain
// ================================================================================

// newPaymentUseCases creates all payment-related use cases
func newPaymentUseCases(
	repos PaymentRepositories,
	memberRepo memberdomain.Repository,
	paymentGateway interface {
		paymentdomain.Gateway
		paymentdomain.GatewayConfig
	},
	validator Validator,
) (PaymentUseCases, SavedCardUseCases, ReceiptUseCases) {
	// Create domain service
	paymentService := paymentdomain.NewService()

	// Special case: Create HandleCallback first since it's needed by ProcessCallbackRetries
	handleCallbackUC := paymentservice.NewHandleCallbackUseCase(repos.Payment, paymentService)

	paymentUseCases := PaymentUseCases{
		InitiatePayment:        paymentservice.NewInitiatePaymentUseCase(repos.Payment, paymentService, paymentGateway, validator),
		VerifyPayment:          paymentservice.NewVerifyPaymentUseCase(repos.Payment, paymentService, paymentGateway),
		HandleCallback:         handleCallbackUC,
		ListMemberPayments:     paymentservice.NewListMemberPaymentsUseCase(repos.Payment),
		CancelPayment:          paymentservice.NewCancelPaymentUseCase(repos.Payment, paymentService),
		RefundPayment:          paymentservice.NewRefundPaymentUseCase(repos.Payment, paymentService, paymentGateway),
		PayWithSavedCard:       savedcardservice.NewPayWithSavedCardUseCase(repos.Payment, repos.SavedCard, paymentService, paymentGateway),
		ExpirePayments:         paymentservice.NewExpirePaymentsUseCase(repos.Payment, paymentService),
		ProcessCallbackRetries: paymentservice.NewProcessCallbackRetriesUseCase(repos.CallbackRetry, handleCallbackUC),
	}

	savedCardUseCases := SavedCardUseCases{
		SaveCard:        paymentservice.NewSaveCardUseCase(repos.SavedCard),
		ListSavedCards:  savedcardservice.NewListSavedCardsUseCase(repos.SavedCard),
		DeleteSavedCard: savedcardservice.NewDeleteSavedCardUseCase(repos.SavedCard),
		SetDefaultCard:  paymentservice.NewSetDefaultCardUseCase(repos.SavedCard),
	}

	receiptUseCases := ReceiptUseCases{
		GenerateReceipt: receiptservice.NewGenerateReceiptUseCase(repos.Payment, repos.Receipt, memberRepo),
		GetReceipt:      receiptservice.NewGetReceiptUseCase(repos.Receipt),
		ListReceipts:    receiptservice.NewListReceiptsUseCase(repos.Receipt),
	}

	return paymentUseCases, savedCardUseCases, receiptUseCases
}
