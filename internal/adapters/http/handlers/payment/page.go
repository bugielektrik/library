package payment

import (
	"html/template"
	"net/http"
	"path/filepath"

	"go.uber.org/zap"

	"library-service/pkg/httputil"
	"library-service/pkg/logutil"
)

// PaymentPageHandler handles serving the payment HTML page.
type PaymentPageHandler struct {
	template *template.Template
}

// NewPaymentPageHandler creates a new payment page handler.
func NewPaymentPageHandler() (*PaymentPageHandler, error) {
	// Load payment template
	tmplPath := filepath.Join("web", "templates", "payment.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, err
	}

	return &PaymentPageHandler{
		template: tmpl,
	}, nil
}

// ServePaymentPage serves the payment HTML page.
// @Summary Payment page
// @Description Displays the payment page for completing a payment
// @Tags payments
// @Produce html
// @Param paymentId query string true "Payment ID"
// @Param invoiceId query string true "Invoice ID"
// @Param authToken query string true "Payment gateway auth token"
// @Param terminal query string true "Terminal ID"
// @Param amount query string true "Amount in cents"
// @Param currency query string false "Currency code" default(KZT)
// @Param backLink query string false "Redirect URL after payment"
// @Param postLink query string false "Callback URL"
// @Success 200 {string} string "HTML page"
// @Router /payment [get]
func (h *PaymentPageHandler) ServePaymentPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "payment_page_handler", "show")

	// Set content type
	w.Header().Set(httputil.HeaderContentType, httputil.ContentTypeHTML)

	// Execute template
	if err := h.template.Execute(w, nil); err != nil {
		logger.Error("failed to render payment template", zap.Error(err))
		http.Error(w, "Failed to load payment page", http.StatusInternalServerError)
		return
	}
}
