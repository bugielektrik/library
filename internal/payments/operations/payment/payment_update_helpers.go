package payment

import (
	"library-service/internal/payments/domain"
)

// UpdatePaymentFromGatewayResponse updates payment entity fields from a gateway response.
// This is a common pattern used across multiple payment use cases.
func UpdatePaymentFromGatewayResponse(
	paymentEntity *domain.Payment,
	transactionID string,
	approvalCode string,
	errorCode string,
	errorMessage string,
) {
	if transactionID != "" {
		paymentEntity.GatewayTransactionID = &transactionID
	}

	if approvalCode != "" {
		paymentEntity.ApprovalCode = &approvalCode
	}

	if errorCode != "" {
		paymentEntity.ErrorCode = &errorCode
	}

	if errorMessage != "" {
		paymentEntity.ErrorMessage = &errorMessage
	}
}

// UpdatePaymentFromCardCharge updates payment entity from a card charge response.
func UpdatePaymentFromCardCharge(
	paymentEntity *domain.Payment,
	gatewayResp *domain.CardChargeResponse,
	paymentService *domain.Service,
) {
	UpdatePaymentFromGatewayResponse(
		paymentEntity,
		gatewayResp.TransactionID,
		gatewayResp.ApprovalCode,
		gatewayResp.ErrorCode,
		gatewayResp.ErrorMessage,
	)

	paymentEntity.Status = paymentService.MapGatewayStatus(gatewayResp.Status)
}
