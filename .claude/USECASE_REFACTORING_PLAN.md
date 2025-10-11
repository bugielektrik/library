# Use Case Refactoring Plan

**Date:** October 11, 2025
**Status:** 🔄 **IN PROGRESS**

## Overview

This document outlines the refactoring plan to bring all use cases into a unified pattern as defined in [USECASE_PATTERN_STANDARDS.md](./.claude/USECASE_PATTERN_STANDARDS.md).

---

## 🎯 Refactoring Goals

1. **Standardize response types** - All use cases return value types, not pointers
2. **Add missing response DTOs** - Delete/Update operations return response DTOs
3. **Add missing request DTOs** - List operations have request DTOs (even if empty)
4. **Consistent Execute signature** - All use cases: `Execute(ctx, req) (response, error)`
5. **Unified error handling** - Consistent error wrapping patterns
6. **Complete documentation** - All use cases have comprehensive comments

---

## 📊 Current State Analysis

### Issues Identified

| Issue | Files Affected | Priority |
|-------|---------------|----------|
| **Pointer return types** | 4 auth use cases | HIGH |
| **Missing response DTOs** | 2 book use cases | HIGH |
| **No request parameter** | 2 list use cases | MEDIUM |
| **Inconsistent documentation** | ~50% of use cases | LOW |
| **Mixed error patterns** | All use cases | LOW |

---

## 🔧 Phase 1: Fix Auth Use Cases (Pointer Returns)

### Files to Modify

#### 1. `internal/usecase/authops/login.go`

**Changes:**
- Change Execute return type: `(*LoginResponse, error)` → `(LoginResponse, error)`
- Change return statement: `return &LoginResponse{...}, nil` → `return LoginResponse{...}, nil`
- Change error returns: `return nil, err` → `return LoginResponse{}, err`

**Impact:**
- HTTP handler: `internal/adapters/http/handlers/auth/login.go`
- Tests: `internal/usecase/authops/login_test.go`

#### 2. `internal/usecase/authops/register.go`

**Changes:**
- Change Execute return type: `(*RegisterResponse, error)` → `(RegisterResponse, error)`
- Change return statement: `return &RegisterResponse{...}, nil` → `return RegisterResponse{...}, nil`
- Change error returns: `return nil, err` → `return RegisterResponse{}, err`

**Impact:**
- HTTP handler: `internal/adapters/http/handlers/auth/register.go`
- Tests: `internal/usecase/authops/register_test.go`

#### 3. `internal/usecase/authops/validate.go`

**Changes:**
- Change Execute return type: `(*ValidateTokenResponse, error)` → `(ValidateTokenResponse, error)`
- Change return statement: `return &ValidateTokenResponse{...}, nil` → `return ValidateTokenResponse{...}, nil`
- Change error returns: `return nil, err` → `return ValidateTokenResponse{}, err`

**Impact:**
- HTTP handler (if any)
- Tests: `internal/usecase/authops/validate_test.go`

#### 4. `internal/usecase/authops/refresh.go`

**Changes:**
- Change Execute return type: `(*RefreshTokenResponse, error)` → `(RefreshTokenResponse, error)`
- Change return statement: `return &RefreshTokenResponse{...}, nil` → `return RefreshTokenResponse{...}, nil`
- Change error returns: `return nil, err` → `return RefreshTokenResponse{}, err`

**Impact:**
- HTTP handler: `internal/adapters/http/handlers/auth/refresh.go`
- Tests: `internal/usecase/authops/refresh_test.go`

---

## 🔧 Phase 2: Add Missing Response DTOs

### Files to Modify

#### 1. `internal/usecase/bookops/delete_book.go`

**Current:**
```go
func (uc *DeleteBookUseCase) Execute(ctx context.Context, req DeleteBookRequest) error
```

**New:**
```go
// DeleteBookResponse represents the output of deleting a book
type DeleteBookResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

func (uc *DeleteBookUseCase) Execute(ctx context.Context, req DeleteBookRequest) (DeleteBookResponse, error)
```

**Changes:**
- Add `DeleteBookResponse` struct
- Change return type from `error` to `(DeleteBookResponse, error)`
- Return `DeleteBookResponse{Success: true, Message: "book deleted successfully"}, nil` on success
- Return `DeleteBookResponse{}, err` on error

**Impact:**
- HTTP handler: `internal/adapters/http/handlers/book/delete.go`
- Tests: `internal/usecase/bookops/delete_book_test.go`

#### 2. `internal/usecase/bookops/update_book.go`

**Current:**
```go
func (uc *UpdateBookUseCase) Execute(ctx context.Context, req UpdateBookRequest) error
```

**New:**
```go
// UpdateBookResponse represents the output of updating a book
type UpdateBookResponse struct {
    Success bool   `json:"success"`
    Message string `json:"message"`
}

func (uc *UpdateBookUseCase) Execute(ctx context.Context, req UpdateBookRequest) (UpdateBookResponse, error)
```

**Changes:**
- Add `UpdateBookResponse` struct
- Change return type from `error` to `(UpdateBookResponse, error)`
- Return `UpdateBookResponse{Success: true, Message: "book updated successfully"}, nil` on success
- Return `UpdateBookResponse{}, err` on error

**Impact:**
- HTTP handler: `internal/adapters/http/handlers/book/update.go`
- Tests: `internal/usecase/bookops/update_book_test.go`

---

## 🔧 Phase 3: Add Missing Request DTOs

### Files to Modify

#### 1. `internal/usecase/authorops/list_authors.go`

**Current:**
```go
func (uc *ListAuthorsUseCase) Execute(ctx context.Context) (ListAuthorsResponse, error)
```

**New:**
```go
// ListAuthorsRequest represents the input for listing authors
type ListAuthorsRequest struct {
    // Future: Add pagination, filters, sorting
}

func (uc *ListAuthorsUseCase) Execute(ctx context.Context, req ListAuthorsRequest) (ListAuthorsResponse, error)
```

**Changes:**
- Add `ListAuthorsRequest` struct (empty for now)
- Add `req` parameter to Execute method signature
- Keep implementation the same

**Impact:**
- HTTP handler: `internal/adapters/http/handlers/author/list.go`
- Container: `internal/usecase/container.go`
- Tests: `internal/usecase/authorops/list_authors_test.go`

#### 2. `internal/usecase/memberops/list_members.go`

**Current:**
```go
func (uc *ListMembersUseCase) Execute(ctx context.Context) (ListMembersResponse, error)
```

**New:**
```go
// ListMembersRequest represents the input for listing members
type ListMembersRequest struct {
    // Future: Add pagination, filters, sorting
}

func (uc *ListMembersUseCase) Execute(ctx context.Context, req ListMembersRequest) (ListMembersResponse, error)
```

**Changes:**
- Add `ListMembersRequest` struct (empty for now)
- Add `req` parameter to Execute method signature
- Keep implementation the same

**Impact:**
- HTTP handler (if any)
- Container: `internal/usecase/container.go`
- Tests: `internal/usecase/memberops/list_members_test.go`

---

## 📝 Phase 4: Update HTTP Handlers

All HTTP handlers that call the modified use cases need to be updated.

### Handlers to Update

#### Auth Handlers
- `internal/adapters/http/handlers/auth/login.go`
  - Change: `response, err := uc.Execute(...)` - no longer needs dereferencing
- `internal/adapters/http/handlers/auth/register.go`
  - Change: `response, err := uc.Execute(...)` - no longer needs dereferencing
- `internal/adapters/http/handlers/auth/refresh.go`
  - Change: `response, err := uc.Execute(...)` - no longer needs dereferencing

#### Book Handlers
- `internal/adapters/http/handlers/book/delete.go`
  - Change: Handle `DeleteBookResponse` instead of just error
- `internal/adapters/http/handlers/book/update.go`
  - Change: Handle `UpdateBookResponse` instead of just error

#### Author Handlers
- `internal/adapters/http/handlers/author/list.go`
  - Change: Pass `ListAuthorsRequest{}` to Execute

#### Member Handlers
- `internal/adapters/http/handlers/member/list.go` (if exists)
  - Change: Pass `ListMembersRequest{}` to Execute

---

## 🧪 Phase 5: Update Tests

### Test Files to Update

#### Auth Tests
- `internal/usecase/authops/login_test.go`
- `internal/usecase/authops/register_test.go`
- `internal/usecase/authops/validate_test.go`
- `internal/usecase/authops/refresh_test.go`

**Changes:**
- Update assertions to expect value types, not pointers
- Update error case assertions to expect empty structs

#### Book Tests
- `internal/usecase/bookops/delete_book_test.go`
- `internal/usecase/bookops/update_book_test.go`

**Changes:**
- Update to expect response DTOs
- Update assertions for success/error cases

#### Author/Member Tests
- `internal/usecase/authorops/list_authors_test.go`
- `internal/usecase/memberops/list_members_test.go`

**Changes:**
- Pass empty request DTOs to Execute

---

## ⚡ Implementation Order

### Step 1: Auth Use Cases (Highest Impact)
1. Update `login.go`, `register.go`, `validate.go`, `refresh.go`
2. Update corresponding HTTP handlers
3. Update tests
4. Run tests and verify

### Step 2: Book Use Cases (Add Responses)
1. Update `delete_book.go`, `update_book.go`
2. Update corresponding HTTP handlers
3. Update tests
4. Run tests and verify

### Step 3: List Use Cases (Add Requests)
1. Update `list_authors.go`, `list_members.go`
2. Update corresponding HTTP handlers (if any)
3. Update tests
4. Run tests and verify

### Step 4: Build and Integration Test
1. Run full build: `make build`
2. Run all tests: `make test`
3. Verify no breaking changes

---

## 📋 Verification Checklist

- [ ] All auth use cases return value types
- [ ] All delete/update use cases return response DTOs
- [ ] All list use cases have request DTOs
- [ ] All HTTP handlers updated
- [ ] All tests updated and passing
- [ ] Build successful
- [ ] No breaking changes to external APIs
- [ ] Documentation updated

---

## 🔍 Automated Changes

### Sed Scripts for Bulk Updates

#### Auth Use Cases - Return Type Changes

```bash
# Login
sed -i '' 's/func (uc \*LoginUseCase) Execute(ctx context.Context, req LoginRequest) (\*LoginResponse, error)/func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (LoginResponse, error)/' internal/usecase/authops/login.go

# Register
sed -i '' 's/func (uc \*RegisterUseCase) Execute(ctx context.Context, req RegisterRequest) (\*RegisterResponse, error)/func (uc *RegisterUseCase) Execute(ctx context.Context, req RegisterRequest) (RegisterResponse, error)/' internal/usecase/authops/register.go

# Validate
sed -i '' 's/func (uc \*ValidateTokenUseCase) Execute(ctx context.Context, req ValidateTokenRequest) (\*ValidateTokenResponse, error)/func (uc *ValidateTokenUseCase) Execute(ctx context.Context, req ValidateTokenRequest) (ValidateTokenResponse, error)/' internal/usecase/authops/validate.go

# Refresh
sed -i '' 's/func (uc \*RefreshTokenUseCase) Execute(ctx context.Context, req RefreshTokenRequest) (\*RefreshTokenResponse, error)/func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req RefreshTokenRequest) (RefreshTokenResponse, error)/' internal/usecase/authops/refresh.go
```

---

## 🎯 Expected Benefits

### Code Consistency
- All use cases follow the same pattern
- Predictable structure across the codebase

### Developer Experience
- Easier to understand and maintain
- Less cognitive load when switching between use cases

### API Consistency
- All endpoints return structured responses
- Better for API consumers

---

## 📊 Risk Assessment

### Low Risk Changes
- ✅ Auth return type changes (internal only)
- ✅ Adding request DTOs for list operations (backward compatible)

### Medium Risk Changes
- ⚠️ Adding response DTOs for delete/update (HTTP handlers need updates)

### Mitigation
- Run full test suite after each phase
- Test HTTP endpoints manually
- Check Swagger docs are still valid

---

## 🔗 Related Documents

1. [USECASE_PATTERN_STANDARDS.md](./.claude/USECASE_PATTERN_STANDARDS.md) - The pattern standard
2. [CODE_PATTERN_STANDARDS.md](./.claude/CODE_PATTERN_STANDARDS.md) - Domain patterns
3. [CODEBASE_PATTERN_REFACTORING.md](./.claude/CODEBASE_PATTERN_REFACTORING.md) - Domain refactoring

---

**Status:** Ready for implementation
**Estimated Time:** 2-3 hours
**Breaking Changes:** None (all changes are internal)
