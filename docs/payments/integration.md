# Payment Integration with epayment.kz

This document provides detailed information about the payment integration using epayment.kz (Halyk Bank's payment gateway) in the Library Management System.

## Overview

The payment system enables library members to make payments for:
- **Fines**: Overdue book fines
- **Subscriptions**: Membership subscription fees
- **Deposits**: Security deposits

## Architecture

The payment feature follows Clean Architecture principles:

```
internal/
├── domain/payment/              # Payment domain (business logic)
│   ├── entity.go               # Payment entity with status enums
│   ├── service.go              # Payment business rules
│   ├── repository.go           # Repository interface
│   └── dto.go                  # Request/Response DTOs
├── usecase/paymentops/         # Payment use cases
│   ├── initiate_payment.go    # Initiate payment flow
│   ├── verify_payment.go      # Check payment status
│   ├── handle_callback.go     # Process gateway callbacks
│   └── list_member_payments.go # List member's payment history
├── adapters/
│   ├── payment/epayment/      # Epayment.kz gateway adapter
│   │   ├── gateway.go         # Gateway implementation
│   │   └── config.go          # Configuration loader
│   ├── repository/postgres/
│   │   └── payment.go         # PostgreSQL repository
│   └── http/
│       ├── handlers/payment.go # HTTP handlers
│       └── dto/payment.go      # HTTP DTOs
```

## Setup

### 1. Environment Configuration

Copy `.env.example` to `.env` and configure epayment.kz credentials:

```bash
# Epayment.kz Payment Configuration (Halyk Bank)
EPAYMENT_ENV=test                    # "test" or "prod"
EPAYMENT_CLIENT_ID=your-client-id
EPAYMENT_CLIENT_SECRET=your-secret
EPAYMENT_TERMINAL=your-terminal-id
EPAYMENT_BACK_LINK=http://localhost:8080/payment/success
EPAYMENT_POST_LINK=http://localhost:8080/api/v1/payments/callback
```

**Important Notes:**
- For production, set `EPAYMENT_ENV=prod`
- Obtain credentials from Halyk Bank merchant portal
- `BACK_LINK`: Where users are redirected after payment
- `POST_LINK`: Where epayment.kz sends callback notifications

### 2. Database Migration

Run the payment migration:

```bash
make migrate-up
```

This creates the `payments` table with proper indexes and constraints.

### 3. Start the Application

```bash
make dev
```

## API Endpoints

### 1. Initiate Payment

**POST** `/api/v1/payments/initiate`

Initiates a new payment and returns payment gateway details.

**Authentication:** Required (JWT Bearer token)

**Request Body:**
```json
{
  "amount": 10000,
  "currency": "KZT",
  "payment_type": "fine",
  "related_entity_id": "uuid-of-fine-or-subscription"
}
```

**Response:**
```json
{
  "payment_id": "payment-uuid",
  "invoice_id": "fine-member-id-timestamp",
  "auth_token": "oauth-token-from-provider",
  "terminal": "terminal-id",
  "amount": 10000,
  "currency": "KZT",
  "back_link": "http://localhost:8080/payment/success",
  "post_link": "http://localhost:8080/api/v1/payments/callback"
}
```

**Frontend Integration:**

Include the epayment.kz JavaScript SDK and call `halyk.pay()`:

```html
<script src="https://test-epay.epayment.kz/redesign-payform/payment-api.js"></script>

<script>
const paymentData = await fetch('/api/v1/payments/initiate', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${accessToken}`,
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    amount: 10000,
    currency: 'KZT',
    payment_type: 'fine'
  })
});

const payment = await paymentData.json();

halyk.pay({
  invoiceId: payment.invoice_id,
  terminal: payment.terminal,
  amount: payment.amount,
  currency: payment.currency,
  auth: payment.auth_token,
  backLink: payment.back_link,
  postLink: payment.post_link
});
</script>
```

### 2. Verify Payment

**GET** `/api/v1/payments/{id}`

Check the current status of a payment.

**Authentication:** Required

**Response:**
```json
{
  "id": "payment-uuid",
  "invoice_id": "fine-member-id-timestamp",
  "status": "completed",
  "amount": 10000,
  "currency": "KZT",
  "gateway_transaction_id": "provider-tx-id",
  "card_mask": "****1234",
  "approval_code": "123456"
}
```

### 3. Payment Callback (Webhook)

**POST** `/api/v1/payments/callback`

**Authentication:** Not required (called by payment gateway)

Receives callbacks from epayment.kz gateway when payment status changes.

**Request Body:**
```json
{
  "invoiceId": "fine-member-id-timestamp",
  "transactionId": "provider-transaction-id",
  "amount": 10000,
  "currency": "KZT",
  "status": "success",
  "cardMask": "****1234",
  "approvalCode": "123456"
}
```

### 4. List Member Payments

**GET** `/api/v1/payments/member/{memberId}`

Get all payments for a specific member.

**Authentication:** Required

**Response:**
```json
{
  "payments": [
    {
      "id": "payment-uuid",
      "invoice_id": "fine-member-id-timestamp",
      "amount": 10000,
      "currency": "KZT",
      "status": "completed",
      "payment_type": "fine",
      "created_at": "2025-01-06T12:00:00Z",
      "completed_at": "2025-01-06T12:05:00Z"
    }
  ]
}
```

## Payment Statuses

- **pending**: Payment created, waiting for user action
- **processing**: Payment is being processed by gateway
- **completed**: Payment successfully completed
- **failed**: Payment failed
- **cancelled**: Payment cancelled by user or system
- **refunded**: Payment refunded

## Status Transitions

Valid status transitions:
- `pending` → `processing`, `cancelled`, `failed`
- `processing` → `completed`, `failed`, `cancelled`
- `completed` → `refunded`
- `failed` → `pending` (retry)

## Payment Types

- **fine**: Payment for overdue book fines
- **subscription**: Membership subscription payment
- **deposit**: Security deposit payment

## Business Logic

### Payment Expiration

Payments automatically expire 30 minutes after creation if not completed.

### Currency Support

Currently supported currencies:
- **KZT** (Kazakhstani Tenge)
- **USD** (US Dollar)
- **EUR** (Euro)
- **RUB** (Russian Ruble)

### Invoice ID Generation

Invoice IDs follow the pattern: `{payment_type}-{member_id}-{timestamp}`

Example: `fine-uuid-1704537600`

## Testing

### Run Payment Tests

```bash
# Unit tests
go test -v ./internal/domain/payment/...
go test -v ./internal/usecase/paymentops/...

# Integration tests
make test-integration
```

### Testing with epayment.kz Test Environment

1. Set `EPAYMENT_ENV=test` in `.env`
2. Use test credentials provided by Halyk Bank
3. Test cards:
   - **Success**: 4405639999999999
   - **Decline**: 4111111111111111

## Security Considerations

1. **Always use HTTPS** in production for payment callbacks
2. **Validate callback signatures** (implement signature verification)
3. **Never log** sensitive payment data (card numbers, CVV)
4. **Store minimal** payment gateway responses
5. **Use environment variables** for all credentials

## Error Handling

Payment errors are mapped to appropriate HTTP status codes:

- **400 Bad Request**: Invalid payment data
- **404 Not Found**: Payment not found
- **409 Conflict**: Payment already processed
- **410 Gone**: Payment expired
- **502 Bad Gateway**: Payment gateway error

## Monitoring

Monitor these metrics:
- Payment success rate
- Average payment processing time
- Failed payment reasons
- Gateway response times

## Troubleshooting

### Gateway Authentication Fails

**Symptom:** `failed to get auth token`

**Solution:**
- Verify `EPAYMENT_CLIENT_ID` and `EPAYMENT_CLIENT_SECRET`
- Check environment (`test` vs `prod`)
- Ensure credentials are active

### Callback Not Received

**Symptom:** Payment stuck in `processing` status

**Solution:**
- Verify `EPAYMENT_POST_LINK` is publicly accessible
- Check firewall/network settings
- Review epayment.kz merchant portal logs

### Payment Expires Quickly

**Symptom:** Payments expire before completion

**Solution:**
- Increase expiration time in `internal/domain/payment/entity.go`:
  ```go
  expiresAt := now.Add(60 * time.Minute) // Increase to 60 minutes
  ```

## Production Checklist

- [ ] Change `EPAYMENT_ENV` to `prod`
- [ ] Use production credentials from Halyk Bank
- [ ] Update `EPAYMENT_BACK_LINK` to production URL
- [ ] Update `EPAYMENT_POST_LINK` to production callback URL
- [ ] Enable HTTPS for all endpoints
- [ ] Implement callback signature verification
- [ ] Set up monitoring and alerting
- [ ] Test with real payment cards
- [ ] Review and configure refund policy

## Support

For epayment.kz integration support:
- **Merchant Portal**: https://epay.homebank.kz/
- **Documentation**: https://epayment.kz/en-US/docs/
- **Technical Support**: Contact Halyk Bank merchant support
