package usecase

import (
	memberdomain "library-service/internal/members/domain"
	paymentdomain "library-service/internal/payments/domain"
	paymentops "library-service/internal/payments/operations/payment"
	receiptops "library-service/internal/payments/operations/receipt"
	savedcardops "library-service/internal/payments/operations/savedcard"
)

// PaymentUseCases contains all payment-related use cases
type PaymentUseCases struct {
	InitiatePayment        *paymentops.InitiatePaymentUseCase
	VerifyPayment          *paymentops.VerifyPaymentUseCase
	HandleCallback         *paymentops.HandleCallbackUseCase
	ListMemberPayments     *paymentops.ListMemberPaymentsUseCase
	CancelPayment          *paymentops.CancelPaymentUseCase
	RefundPayment          *paymentops.RefundPaymentUseCase
	PayWithSavedCard       *savedcardops.PayWithSavedCardUseCase
	ExpirePayments         *paymentops.ExpirePaymentsUseCase
	ProcessCallbackRetries *paymentops.ProcessCallbackRetriesUseCase
}

// SavedCardUseCases contains all saved card-related use cases
type SavedCardUseCases struct {
	SaveCard        *paymentops.SaveCardUseCase
	ListSavedCards  *savedcardops.ListSavedCardsUseCase
	DeleteSavedCard *savedcardops.DeleteSavedCardUseCase
	SetDefaultCard  *paymentops.SetDefaultCardUseCase
}

// ReceiptUseCases contains all receipt-related use cases
type ReceiptUseCases struct {
	GenerateReceipt *receiptops.GenerateReceiptUseCase
	GetReceipt      *receiptops.GetReceiptUseCase
	ListReceipts    *receiptops.ListReceiptsUseCase
}

// PaymentRepositories contains payment-related repositories
type PaymentRepositories struct {
	Payment       paymentdomain.Repository
	SavedCard     paymentdomain.SavedCardRepository
	CallbackRetry paymentdomain.CallbackRetryRepository
	Receipt       paymentdomain.ReceiptRepository
}

// newPaymentUseCases creates all payment-related use cases
func newPaymentUseCases(
	repos PaymentRepositories,
	memberRepo memberdomain.Repository,
	paymentGateway interface {
		paymentdomain.Gateway
		paymentdomain.GatewayConfig
	},
) (PaymentUseCases, SavedCardUseCases, ReceiptUseCases) {
	// Create domain service
	paymentService := paymentdomain.NewService()

	// Special case: Create HandleCallback first since it's needed by ProcessCallbackRetries
	handleCallbackUC := paymentops.NewHandleCallbackUseCase(repos.Payment, paymentService)

	paymentUseCases := PaymentUseCases{
		InitiatePayment:        paymentops.NewInitiatePaymentUseCase(repos.Payment, paymentService, paymentGateway),
		VerifyPayment:          paymentops.NewVerifyPaymentUseCase(repos.Payment, paymentService, paymentGateway),
		HandleCallback:         handleCallbackUC,
		ListMemberPayments:     paymentops.NewListMemberPaymentsUseCase(repos.Payment),
		CancelPayment:          paymentops.NewCancelPaymentUseCase(repos.Payment, paymentService),
		RefundPayment:          paymentops.NewRefundPaymentUseCase(repos.Payment, paymentService, paymentGateway),
		PayWithSavedCard:       savedcardops.NewPayWithSavedCardUseCase(repos.Payment, repos.SavedCard, paymentService, paymentGateway),
		ExpirePayments:         paymentops.NewExpirePaymentsUseCase(repos.Payment, paymentService),
		ProcessCallbackRetries: paymentops.NewProcessCallbackRetriesUseCase(repos.CallbackRetry, handleCallbackUC),
	}

	savedCardUseCases := SavedCardUseCases{
		SaveCard:        paymentops.NewSaveCardUseCase(repos.SavedCard),
		ListSavedCards:  savedcardops.NewListSavedCardsUseCase(repos.SavedCard),
		DeleteSavedCard: savedcardops.NewDeleteSavedCardUseCase(repos.SavedCard),
		SetDefaultCard:  paymentops.NewSetDefaultCardUseCase(repos.SavedCard),
	}

	receiptUseCases := ReceiptUseCases{
		GenerateReceipt: receiptops.NewGenerateReceiptUseCase(repos.Payment, repos.Receipt, memberRepo),
		GetReceipt:      receiptops.NewGetReceiptUseCase(repos.Receipt),
		ListReceipts:    receiptops.NewListReceiptsUseCase(repos.Receipt),
	}

	return paymentUseCases, savedCardUseCases, receiptUseCases
}
