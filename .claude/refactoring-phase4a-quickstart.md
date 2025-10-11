# Phase 4A Quick Start: Test Modernization

## 🚀 Ready to Start Now

### Step 1: Update First Test File (5 minutes)

Pick `internal/usecase/authops/login_test.go`:

```go
// REMOVE these lines:
type mockMemberRepository struct {
    mock.Mock
}

func (m *mockMemberRepository) GetByEmail(ctx context.Context, email string) (member.Member, error) {
    args := m.Called(ctx, email)
    return args.Get(0).(member.Member), args.Error(1)
}

// ADD these lines:
import "library-service/internal/adapters/repository/mocks"

// In test function:
mockRepo := new(mocks.MockMemberRepository)
mockRepo.On("GetByEmail", ctx, "test@example.com").Return(testMember, nil)
```

### Step 2: Create First Test Builder (10 minutes)

Create `test/builders/member_builder.go`:

```go
package builders

import (
    "time"
    "library-service/internal/domain/member"
)

type MemberBuilder struct {
    member member.Member
}

func Member() *MemberBuilder {
    return &MemberBuilder{
        member: member.Member{
            ID:        "test-member-id",
            Email:     "test@example.com",
            FullName:  "Test User",
            Role:      member.RoleUser,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }
}

func (b *MemberBuilder) WithID(id string) *MemberBuilder {
    b.member.ID = id
    return b
}

func (b *MemberBuilder) WithEmail(email string) *MemberBuilder {
    b.member.Email = email
    return b
}

func (b *MemberBuilder) AsAdmin() *MemberBuilder {
    b.member.Role = member.RoleAdmin
    return b
}

func (b *MemberBuilder) WithSubscription(tier member.SubscriptionTier) *MemberBuilder {
    b.member.SubscriptionTier = &tier
    expiry := time.Now().Add(30 * 24 * time.Hour)
    b.member.SubscriptionExpiry = &expiry
    return b
}

func (b *MemberBuilder) Build() member.Member {
    return b.member
}

// Usage in tests:
// testMember := builders.Member().WithEmail("custom@test.com").AsAdmin().Build()
```

### Step 3: Create Test Helper (5 minutes)

Create `test/helpers/assertions.go`:

```go
package helpers

import (
    "testing"
    "reflect"
)

func AssertEqual(t *testing.T, expected, actual interface{}) {
    t.Helper()
    if !reflect.DeepEqual(expected, actual) {
        t.Errorf("Expected %v, got %v", expected, actual)
    }
}

func AssertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
}

func AssertError(t *testing.T, err error) {
    t.Helper()
    if err == nil {
        t.Fatal("Expected error but got nil")
    }
}

func AssertErrorContains(t *testing.T, err error, substring string) {
    t.Helper()
    if err == nil {
        t.Fatal("Expected error but got nil")
    }
    if !strings.Contains(err.Error(), substring) {
        t.Errorf("Error %q does not contain %q", err.Error(), substring)
    }
}
```

## 📋 Complete File List

### Test Files to Update (18 files)

```bash
# Auth tests (3 files)
internal/usecase/authops/validate_test.go
internal/usecase/authops/login_test.go
internal/usecase/authops/refresh_test.go

# Member tests (3 files)
internal/usecase/memberops/list_members_test.go
internal/usecase/memberops/get_member_profile_test.go
internal/usecase/memberops/helpers_test.go

# Payment tests (10 files)
internal/usecase/paymentops/process_callback_retries_test.go
internal/usecase/paymentops/generate_receipt_test.go
internal/usecase/paymentops/pay_with_saved_card_test.go
internal/usecase/paymentops/refund_payment_test.go
internal/usecase/paymentops/cancel_payment_test.go
internal/usecase/paymentops/list_member_payments_test.go
internal/usecase/paymentops/expire_payments_test.go
internal/usecase/paymentops/verify_payment_test.go
internal/usecase/paymentops/initiate_payment_test.go
internal/usecase/paymentops/handle_callback_test.go

# Other tests (2 files)
internal/usecase/subops/subscribe_member_test.go
internal/usecase/reservationops/create_reservation_test.go
```

## 🎯 Quick Wins (Do These First!)

### 1. Smallest Test File First
Start with `internal/usecase/memberops/helpers_test.go` - likely the smallest

### 2. Use Find & Replace
```bash
# Find all mock definitions
grep -r "type mock.*Repository struct" internal/usecase --include="*_test.go"

# Replace pattern:
# OLD: type mockMemberRepository struct
# NEW: mockRepo := new(mocks.MockMemberRepository)
```

### 3. Batch Import Addition
Add to all test files at once:
```go
import (
    "library-service/internal/adapters/repository/mocks"
    "library-service/test/builders"
    "library-service/test/helpers"
)
```

## 🔧 Automation Scripts

### Script to Update Imports
```bash
#!/bin/bash
# update-test-imports.sh

for file in $(find internal/usecase -name "*_test.go"); do
    # Add import if not present
    if ! grep -q "internal/adapters/repository/mocks" "$file"; then
        sed -i '' '/^import (/a\
    "library-service/internal/adapters/repository/mocks"
' "$file"
    fi
done
```

### Script to Find Mock Definitions
```bash
#!/bin/bash
# find-old-mocks.sh

echo "Files with old mock patterns:"
grep -l "type mock.*Repository struct" internal/usecase/**/*_test.go

echo -e "\nTotal files to update:"
grep -l "type mock.*Repository struct" internal/usecase/**/*_test.go | wc -l
```

## ✅ Success Checklist

For each test file:
- [ ] Remove old mock struct definitions
- [ ] Import centralized mocks package
- [ ] Replace mock instantiation
- [ ] Update mock method calls
- [ ] Verify tests still pass
- [ ] Use builders where applicable
- [ ] Apply test helpers for assertions

## 💡 Pro Tips

1. **Start Small**: Update one file completely before moving to others
2. **Run Tests Often**: `go test ./...` after each file update
3. **Use Builders Gradually**: Don't feel obligated to use builders everywhere immediately
4. **Keep Old Tests Running**: Ensure backward compatibility during migration
5. **Commit Often**: Small, focused commits make rollback easier

## 🚦 Ready to Start?

1. Run existing tests to ensure baseline: `make test`
2. Pick the smallest test file
3. Apply the pattern
4. Run tests again
5. Commit with message: "refactor: update [filename] to use centralized mocks"
6. Repeat!

---

**Estimated Time**:
- Per file: 5-10 minutes
- Total Phase 4A: 3-4 hours
- Can be done incrementally over several days