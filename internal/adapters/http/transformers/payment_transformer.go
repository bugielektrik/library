package transformers

import (
	"library-service/internal/adapters/http/dto"
	paymentops "library-service/internal/payments/operations/payment"
)

// PaymentTransformer handles payment domain to DTO transformations
type PaymentTransformer struct{}

// NewPaymentTransformer creates a new payment transformer
func NewPaymentTransformer() *PaymentTransformer {
	return &PaymentTransformer{}
}

// ToInitiatePaymentResponse converts use case response to DTO
func (t *PaymentTransformer) ToInitiatePaymentResponse(res paymentops.InitiatePaymentResponse) dto.InitiatePaymentResponse {
	return dto.ToInitiatePaymentResponse(res)
}

// ToVerifyPaymentResponse converts use case response to DTO
func (t *PaymentTransformer) ToVerifyPaymentResponse(res paymentops.VerifyPaymentResponse) dto.PaymentResponse {
	return dto.ToPaymentResponse(res)
}

// ToPaymentListResponse converts use case response to DTO list
func (t *PaymentTransformer) ToPaymentListResponse(res paymentops.ListMemberPaymentsResponse) []dto.PaymentSummaryResponse {
	return dto.ToPaymentSummaryResponses(res.Payments)
}
