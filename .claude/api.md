# API Documentation

> **REST API endpoints, authentication, and design standards**

## Base URL

```
Development: http://localhost:8080
Production:  https://api.library.example.com
```

## API Versioning

All endpoints are prefixed with `/api/v1/`

## Authentication

### JWT Bearer Token

Protected endpoints require JWT token in Authorization header:

```bash
Authorization: Bearer <access_token>
```

### Authentication Flow

**1. Register User**
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "full_name": "John Doe"
}

# Response: 201 Created
{
  "member": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "full_name": "John Doe",
    "role": "user",
    "subscription_type": "free",
    "created_at": "2025-01-15T10:30:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2025-01-16T10:30:00Z"
  }
}
```

**2. Login**
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

# Response: 200 OK
{
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2025-01-16T10:30:00Z"
  },
  "member": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "role": "user"
  }
}
```

**3. Refresh Token**
```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}

# Response: 200 OK
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_at": "2025-01-16T11:30:00Z"
}
```

## Endpoints

### Health Check

```bash
GET /health

# Response: 200 OK
{
  "status": "ok",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### Books

**List Books**
```bash
GET /api/v1/books?page=1&limit=20&genre=Programming

# Response: 200 OK
{
  "books": [
    {
      "id": "book-id-1",
      "name": "Clean Code",
      "isbn": "9780132350884",
      "genre": "Programming",
      "publication_year": 2008,
      "total_copies": 10,
      "available_copies": 7,
      "authors": ["author-id-1"],
      "created_at": "2025-01-10T12:00:00Z",
      "updated_at": "2025-01-10T12:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 156,
    "total_pages": 8
  }
}
```

**Get Book**
```bash
GET /api/v1/books/:id

# Response: 200 OK
{
  "id": "book-id-1",
  "name": "Clean Code",
  "isbn": "9780132350884",
  "genre": "Programming",
  "publication_year": 2008,
  "total_copies": 10,
  "available_copies": 7,
  "authors": [
    {
      "id": "author-id-1",
      "name": "Robert C. Martin",
      "bio": "Software craftsman and author"
    }
  ],
  "created_at": "2025-01-10T12:00:00Z",
  "updated_at": "2025-01-10T12:00:00Z"
}
```

**Create Book** (Requires Authentication)
```bash
POST /api/v1/books
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Clean Code",
  "isbn": "9780132350884",
  "genre": "Programming",
  "publication_year": 2008,
  "total_copies": 10,
  "authors": ["author-id-1"]
}

# Response: 201 Created
{
  "id": "book-id-1",
  "name": "Clean Code",
  "isbn": "9780132350884",
  "genre": "Programming",
  "publication_year": 2008,
  "total_copies": 10,
  "available_copies": 10,
  "authors": ["author-id-1"],
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

**Update Book** (Requires Authentication)
```bash
PUT /api/v1/books/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Clean Code: A Handbook",
  "total_copies": 15
}

# Response: 200 OK
{
  "id": "book-id-1",
  "name": "Clean Code: A Handbook",
  "isbn": "9780132350884",
  "total_copies": 15,
  "updated_at": "2025-01-15T11:00:00Z"
}
```

**Delete Book** (Requires Authentication)
```bash
DELETE /api/v1/books/:id
Authorization: Bearer <token>

# Response: 204 No Content
```

### Members

**Subscribe Member** (Requires Authentication)
```bash
POST /api/v1/members/:id/subscribe
Authorization: Bearer <token>
Content-Type: application/json

{
  "subscription_type": "premium",
  "months": 12
}

# Response: 200 OK
{
  "member_id": "member-id-1",
  "subscription_type": "premium",
  "price": 95.88,
  "expires_at": "2026-01-15T10:30:00Z",
  "discount_percentage": 20
}
```

## Error Responses

All errors follow a consistent format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid ISBN format",
    "details": {
      "field": "isbn",
      "value": "invalid-isbn"
    }
  }
}
```

### Error Codes

| HTTP Status | Error Code | Description |
|------------|------------|-------------|
| 400 | `VALIDATION_ERROR` | Invalid input data |
| 401 | `UNAUTHORIZED` | Missing or invalid authentication |
| 403 | `FORBIDDEN` | Insufficient permissions |
| 404 | `NOT_FOUND` | Resource not found |
| 409 | `CONFLICT` | Resource already exists |
| 422 | `UNPROCESSABLE_ENTITY` | Valid syntax but semantic errors |
| 500 | `INTERNAL_ERROR` | Server error |

### Example Errors

**Validation Error**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request data",
    "details": {
      "isbn": "must be valid ISBN-10 or ISBN-13",
      "total_copies": "must be greater than 0"
    }
  }
}
```

**Unauthorized**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired token"
  }
}
```

**Not Found**
```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Book not found",
    "details": {
      "resource": "book",
      "id": "invalid-id"
    }
  }
}
```

## Request/Response Standards

### Pagination

All list endpoints support pagination:

```bash
GET /api/v1/books?page=1&limit=20
```

**Query Parameters:**
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

**Response:**
```json
{
  "books": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 156,
    "total_pages": 8
  }
}
```

### Filtering

```bash
GET /api/v1/books?genre=Programming&year=2008
```

### Sorting

```bash
GET /api/v1/books?sort=name&order=asc
```

**Query Parameters:**
- `sort` - Field to sort by
- `order` - `asc` or `desc` (default: asc)

### Date Format

All dates use ISO 8601 format with timezone:

```json
{
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-15T10:30:00Z"
}
```

### UUIDs

All resource IDs use UUID v4 format:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Rate Limiting

- **Limit:** 100 requests per minute per IP
- **Headers:**
  - `X-RateLimit-Limit: 100`
  - `X-RateLimit-Remaining: 95`
  - `X-RateLimit-Reset: 1642248000`

**Rate Limit Exceeded:**
```json
HTTP/1.1 429 Too Many Requests
Retry-After: 60

{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests, please try again later"
  }
}
```

## CORS

**Allowed Origins:** Configured via environment variable

**Headers:**
- `Access-Control-Allow-Origin: *` (development)
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

## Testing with cURL

### Quick Test Script

```bash
#!/bin/bash

# Register and login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')

echo "Token: $TOKEN"

# Create book
curl -X POST http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Clean Code",
    "isbn": "9780132350884",
    "genre": "Programming",
    "publication_year": 2008,
    "total_copies": 10,
    "authors": []
  }' | jq

# List books
curl http://localhost:8080/api/v1/books | jq
```

## OpenAPI/Swagger

**Swagger UI:** http://localhost:8080/swagger/index.html

**OpenAPI Spec:** http://localhost:8080/swagger/doc.json

### Generate Swagger Docs

```bash
make gen-docs

# Or manually
swag init -g cmd/api/main.go -o docs
```

## Postman Collection

Import the Postman collection from `api/postman/library-api.json`

**Environment Variables:**
- `base_url`: http://localhost:8080
- `access_token`: (auto-set from login response)

## Next Steps

- Review [Development Guide](./development.md) for API testing workflow
- Check [Testing Guide](./testing.md) for HTTP handler tests
- See [Architecture Guide](./architecture.md) for implementation details
