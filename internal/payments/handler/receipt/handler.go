package receipt

import (
	"library-service/internal/pkg/handlers"
	"library-service/internal/pkg/httputil"
	"library-service/internal/pkg/logutil"
	"library-service/internal/pkg/middleware"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"library-service/internal/container"
	receiptops "library-service/internal/payments/service/receipt"
)

// ReceiptHandler handles HTTP requests for receipts
type ReceiptHandler struct {
	handlers.BaseHandler
	useCases  *container.Container
	validator *middleware.Validator
}

// NewReceiptHandler creates a new receipt handler
func NewReceiptHandler(
	useCases *container.Container,
	validator *middleware.Validator,
) *ReceiptHandler {
	return &ReceiptHandler{
		useCases:  useCases,
		validator: validator,
	}
}

// Routes returns the router for receipt endpoints
func (h *ReceiptHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.generateReceipt)
	r.Get("/", h.listReceipts)
	r.Get("/{id}", h.getReceipt)

	return r
}

// @Summary Generate receipt for payment
// @Description Generates a receipt for a completed payment
// @Tags receipts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body GenerateReceiptRequest true "Receipt generation request"
// @Success 201 {object} ReceiptResponse "Receipt generated"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Payment not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /receipts [post]
func (h *ReceiptHandler) generateReceipt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "receipt_handler", "generate")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Decode request
	var req GenerateReceiptRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute use case
	result, err := h.useCases.Receipt.GenerateReceipt.Execute(ctx, receiptops.GenerateReceiptRequest{
		PaymentID: req.PaymentID,
		MemberID:  memberID,
		Notes:     req.Notes,
	})

	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Get full receipt for response
	fullReceipt, err := h.useCases.Receipt.GetReceipt.Execute(ctx, receiptops.GetReceiptRequest{
		ReceiptID: result.ReceiptID,
		MemberID:  memberID,
	})

	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("receipt generated",
		zap.String("receipt_id", result.ReceiptID),
		zap.String("receipt_number", result.ReceiptNumber),
	)

	h.RespondJSON(w, http.StatusCreated, FromReceiptEntity(fullReceipt.Receipt))
}

// @Summary Get receipt by ID
// @Description Retrieves a receipt by its ID
// @Tags receipts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Receipt ID"
// @Success 200 {object} ReceiptResponse "Receipt details"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "Receipt not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /receipts/{id} [get]
func (h *ReceiptHandler) getReceipt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "receipt_handler", "get")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Get receipt ID from URL
	receiptID, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute use case
	result, err := h.useCases.Receipt.GetReceipt.Execute(ctx, receiptops.GetReceiptRequest{
		ReceiptID: receiptID,
		MemberID:  memberID,
	})

	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("receipt retrieved", zap.String("receipt_id", receiptID))

	h.RespondJSON(w, http.StatusOK, FromReceiptEntity(result.Receipt))
}

// @Summary List member receipts
// @Description Retrieves all receipts for the authenticated member
// @Tags receipts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ListReceiptsResponse "List of receipts"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /receipts [get]
func (h *ReceiptHandler) listReceipts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "receipt_handler", "list")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Execute use case
	result, err := h.useCases.Receipt.ListReceipts.Execute(ctx, receiptops.ListReceiptsRequest{
		MemberID: memberID,
	})

	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	logger.Info("receipts listed", zap.Int("count", result.Total))

	h.RespondJSON(w, http.StatusOK, ListReceiptsResponse{
		Receipts: FromReceiptEntities(result.Receipts),
		Total:    result.Total,
	})
}
