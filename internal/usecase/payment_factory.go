package usecase

import (
	"library-service/internal/domain/member"
	"library-service/internal/domain/payment"
	"library-service/internal/usecase/paymentops"
)

// PaymentUseCases contains all payment-related use cases
type PaymentUseCases struct {
	InitiatePayment        *paymentops.InitiatePaymentUseCase
	VerifyPayment          *paymentops.VerifyPaymentUseCase
	HandleCallback         *paymentops.HandleCallbackUseCase
	ListMemberPayments     *paymentops.ListMemberPaymentsUseCase
	CancelPayment          *paymentops.CancelPaymentUseCase
	RefundPayment          *paymentops.RefundPaymentUseCase
	PayWithSavedCard       *paymentops.PayWithSavedCardUseCase
	ExpirePayments         *paymentops.ExpirePaymentsUseCase
	ProcessCallbackRetries *paymentops.ProcessCallbackRetriesUseCase
}

// SavedCardUseCases contains all saved card-related use cases
type SavedCardUseCases struct {
	SaveCard        *paymentops.SaveCardUseCase
	ListSavedCards  *paymentops.ListSavedCardsUseCase
	DeleteSavedCard *paymentops.DeleteSavedCardUseCase
	SetDefaultCard  *paymentops.SetDefaultCardUseCase
}

// ReceiptUseCases contains all receipt-related use cases
type ReceiptUseCases struct {
	GenerateReceipt *paymentops.GenerateReceiptUseCase
	GetReceipt      *paymentops.GetReceiptUseCase
	ListReceipts    *paymentops.ListReceiptsUseCase
}

// PaymentRepositories contains payment-related repositories
type PaymentRepositories struct {
	Payment       payment.Repository
	SavedCard     payment.SavedCardRepository
	CallbackRetry payment.CallbackRetryRepository
	Receipt       payment.ReceiptRepository
}

// newPaymentUseCases creates all payment-related use cases
func newPaymentUseCases(
	repos PaymentRepositories,
	memberRepo member.Repository,
	paymentGateway interface {
		payment.Gateway
		payment.GatewayConfig
	},
) (PaymentUseCases, SavedCardUseCases, ReceiptUseCases) {
	// Create domain service
	paymentService := payment.NewService()

	// Special case: Create HandleCallback first since it's needed by ProcessCallbackRetries
	handleCallbackUC := paymentops.NewHandleCallbackUseCase(repos.Payment, paymentService)

	paymentUseCases := PaymentUseCases{
		InitiatePayment:        paymentops.NewInitiatePaymentUseCase(repos.Payment, paymentService, paymentGateway),
		VerifyPayment:          paymentops.NewVerifyPaymentUseCase(repos.Payment, paymentService, paymentGateway),
		HandleCallback:         handleCallbackUC,
		ListMemberPayments:     paymentops.NewListMemberPaymentsUseCase(repos.Payment),
		CancelPayment:          paymentops.NewCancelPaymentUseCase(repos.Payment, paymentService),
		RefundPayment:          paymentops.NewRefundPaymentUseCase(repos.Payment, paymentService, paymentGateway),
		PayWithSavedCard:       paymentops.NewPayWithSavedCardUseCase(repos.Payment, repos.SavedCard, paymentService, paymentGateway),
		ExpirePayments:         paymentops.NewExpirePaymentsUseCase(repos.Payment, paymentService),
		ProcessCallbackRetries: paymentops.NewProcessCallbackRetriesUseCase(repos.CallbackRetry, handleCallbackUC),
	}

	savedCardUseCases := SavedCardUseCases{
		SaveCard:        paymentops.NewSaveCardUseCase(repos.SavedCard),
		ListSavedCards:  paymentops.NewListSavedCardsUseCase(repos.SavedCard),
		DeleteSavedCard: paymentops.NewDeleteSavedCardUseCase(repos.SavedCard),
		SetDefaultCard:  paymentops.NewSetDefaultCardUseCase(repos.SavedCard),
	}

	receiptUseCases := ReceiptUseCases{
		GenerateReceipt: paymentops.NewGenerateReceiptUseCase(repos.Payment, repos.Receipt, memberRepo),
		GetReceipt:      paymentops.NewGetReceiptUseCase(repos.Receipt),
		ListReceipts:    paymentops.NewListReceiptsUseCase(repos.Receipt),
	}

	return paymentUseCases, savedCardUseCases, receiptUseCases
}
