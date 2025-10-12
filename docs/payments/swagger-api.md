# Payment API - Swagger Documentation

This document provides a quick reference for the payment endpoints documented in Swagger.

## Access Swagger UI

Once the server is running, access the interactive Swagger documentation at:

```
http://localhost:8080/swagger/index.html
```

## Payment Endpoints

### 1. Initiate Payment

**POST** `/api/v1/payments/initiate`

Initiates a new payment and returns details needed to process the payment through epayment.kz gateway.

**Authentication:** Required (Bearer Token)

**Request Body:**
```json
{
  "amount": 10000,
  "currency": "KZT",
  "payment_type": "fine",
  "related_entity_id": "uuid-optional"
}
```

**Payment Types:**
- `fine` - Payment for library fines
- `subscription` - Membership subscription payment
- `deposit` - Security deposit

**Supported Currencies:**
- `KZT` - Kazakhstani Tenge
- `USD` - US Dollar
- `EUR` - Euro
- `RUB` - Russian Ruble

**Response (200):**
```json
{
  "payment_id": "uuid",
  "invoice_id": "fine-member-id-timestamp",
  "auth_token": "oauth-token",
  "terminal": "terminal-id",
  "amount": 10000,
  "currency": "KZT",
  "back_link": "http://localhost:8080/payment/success",
  "post_link": "http://localhost:8080/api/v1/payments/callback"
}
```

### 2. Verify Payment

**GET** `/api/v1/payments/{id}`

Retrieves the current status and details of a payment.

**Authentication:** Required (Bearer Token)

**Path Parameters:**
- `id` (string, required) - Payment UUID

**Response (200):**
```json
{
  "id": "payment-uuid",
  "invoice_id": "fine-member-id-timestamp",
  "member_id": "member-uuid",
  "amount": 10000,
  "currency": "KZT",
  "status": "completed",
  "payment_method": "card",
  "payment_type": "fine",
  "related_entity_id": "entity-uuid",
  "gateway_transaction_id": "provider-tx-id",
  "card_mask": "****1234",
  "approval_code": "123456",
  "error_code": null,
  "error_message": null,
  "created_at": "2025-01-06T12:00:00Z",
  "updated_at": "2025-01-06T12:05:00Z",
  "completed_at": "2025-01-06T12:05:00Z",
  "expires_at": "2025-01-06T12:30:00Z"
}
```

**Payment Statuses:**
- `pending` - Payment initiated, awaiting processing
- `processing` - Payment being processed by gateway
- `completed` - Payment successfully completed
- `failed` - Payment failed
- `cancelled` - Payment cancelled
- `refunded` - Payment refunded

**Payment Methods:**
- `card` - Card payment
- `wallet` - E-wallet payment

### 3. Payment Callback (Webhook)

**POST** `/api/v1/payments/callback`

Receives callbacks from the epayment.kz payment gateway when payment status changes.

**Authentication:** NOT Required (called by payment gateway)

**Request Body:**
```json
{
  "invoiceId": "fine-member-id-timestamp",
  "transactionId": "provider-tx-id",
  "amount": 10000,
  "currency": "KZT",
  "status": "success",
  "cardMask": "****1234",
  "approvalCode": "123456",
  "errorCode": null,
  "errorMessage": null,
  "extra": {}
}
```

**Response (200):**
```json
{
  "payment_id": "payment-uuid",
  "status": "completed",
  "message": "Payment callback processed successfully"
}
```

### 4. List Member Payments

**GET** `/api/v1/payments/member/{memberId}`

Retrieves all payments for a specific member.

**Authentication:** Required (Bearer Token)

**Path Parameters:**
- `memberId` (string, required) - Member UUID

**Response (200):**
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

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request
```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "details": {
    "field": "amount",
    "reason": "amount must be greater than 0"
  }
}
```

### 401 Unauthorized
```json
{
  "code": "UNAUTHORIZED",
  "message": "Authentication required"
}
```

### 404 Not Found
```json
{
  "code": "PAYMENT_NOT_FOUND",
  "message": "Payment not found",
  "details": {
    "payment_id": "invalid-uuid"
  }
}
```

### 500 Internal Server Error
```json
{
  "code": "INTERNAL_ERROR",
  "message": "Internal server error"
}
```

### 502 Bad Gateway
```json
{
  "code": "PAYMENT_GATEWAY_ERROR",
  "message": "Payment provider error"
}
```

## Using Swagger UI

### 1. Authenticate

1. Click the **"Authorize"** button at the top right
2. Enter your JWT token in the format: `Bearer your-access-token`
3. Click **"Authorize"**
4. Click **"Close"**

### 2. Test Endpoints

1. Expand the endpoint you want to test
2. Click **"Try it out"**
3. Fill in the required parameters
4. Click **"Execute"**
5. View the response below

### 3. Example Flow

**Step 1: Register/Login to get JWT token**
```bash
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "Test123!@#"
}
```

**Step 2: Authorize in Swagger UI**
- Use the `access_token` from login response

**Step 3: Initiate Payment**
```bash
POST /api/v1/payments/initiate
{
  "amount": 50000,
  "currency": "KZT",
  "payment_type": "fine"
}
```

**Step 4: Use the response to integrate with epayment.kz**
```javascript
halyk.pay({
  invoiceId: response.invoice_id,
  terminal: response.terminal,
  amount: response.amount,
  currency: response.currency,
  auth: response.auth_token,
  backLink: response.back_link,
  postLink: response.post_link
});
```

**Step 5: Check Payment Status**
```bash
GET /api/v1/payments/{payment_id}
```

## Schema Definitions

All request/response schemas are fully documented in Swagger with:
- Field names and types
- Required fields
- Field descriptions
- Validation constraints
- Example values

Browse the **"Schemas"** section at the bottom of Swagger UI for complete details.

## Tips

1. **Authentication**: All endpoints except `/callback` require JWT authentication
2. **Amount Format**: Amounts are in the smallest currency unit (e.g., 10000 = 100.00 KZT)
3. **Callback Endpoint**: Must be publicly accessible for payment gateway
4. **Invoice ID**: Automatically generated in format `{payment_type}-{member_id}-{timestamp}`
5. **Expiration**: Payments expire 30 minutes after creation if not completed

## Generated Files

The Swagger documentation is generated from code annotations and includes:

- `api/openapi/docs.go` - Go code documentation
- `api/openapi/swagger.json` - JSON specification
- `api/openapi/swagger.yaml` - YAML specification

## Regenerating Documentation

To regenerate Swagger docs after code changes:

```bash
make gen-docs
```

Or manually:

```bash
swag init -g cmd/api/main.go -o api/openapi --parseDependency --parseInternal
```
