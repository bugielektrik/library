# Payment Integration Documentation

Complete documentation for the Library Management System payment integration with epayment.kz (Kazakhstan payment gateway).

---

## 📚 Documentation Files

### Quick Start
- **[quick-start.md](./quick-start.md)** - 5-minute getting started guide
  - Configuration setup
  - Test API keys
  - First payment flow
  - Callback testing

### Integration Guide
- **[integration.md](./integration.md)** - Complete integration documentation
  - Architecture overview
  - Payment flow diagrams
  - Use case patterns
  - Error handling
  - Best practices

### Features
- **[features.md](./features.md)** - Detailed feature documentation
  - Payment types (fine, subscription, purchase)
  - Supported currencies (KZT, USD, EUR, RUB)
  - Payment methods (card, saved card)
  - Refunds (full & partial)
  - Receipt generation
  - Background worker processes

### API Integration
- **[api-integration.md](./api-integration.md)** - epayment.kz API reference
  - Authentication
  - Endpoints
  - Request/response formats
  - Webhook callbacks
  - Error codes

### Swagger API
- **[swagger-api.md](./swagger-api.md)** - OpenAPI/Swagger annotations
  - API documentation patterns
  - Testing with Swagger UI
  - Authentication in Swagger

---

## 🔧 Environment Setup

```bash
# Required environment variables
EPAYMENT_BASE_URL="https://api.epayment.kz"
EPAYMENT_CLIENT_ID="your-client-id"
EPAYMENT_CLIENT_SECRET="your-client-secret"
EPAYMENT_TERMINAL="your-terminal-id"
```

---

## 🚀 Quick Links

- **Swagger UI:** http://localhost:8080/swagger/index.html (when server running)
- **Payment Domain:** `internal/payments/domain/`
- **Payment Operations:** `internal/payments/operations/payment/`
- **Payment Gateway:** `internal/payments/gateway/epayment/`
- **Worker Process:** `cmd/worker/` (handles payment expiry + callbacks)

---

## 📋 Payment Flow Summary

```
1. Member initiates payment → POST /api/v1/payments/initiate
2. System creates payment record (PENDING status)
3. Member redirected to epayment.kz gateway
4. Member completes payment
5. Gateway sends webhook → POST /api/v1/payments/callback
6. System verifies payment → Updates status to COMPLETED
7. Receipt auto-generated (RCP-YYYY-NNNNNN format)
8. Member can download receipt
```

---

## 🧪 Testing

See [quick-start.md](./quick-start.md) for test credentials and callback testing instructions.

---

**Last Updated:** October 11, 2025
