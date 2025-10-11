#!/bin/bash
# Script to update test files to use centralized mocks

echo "Updating test files to use centralized mocks..."

# List of test files to update
test_files=(
    "internal/usecase/authops/validate_test.go"
    "internal/usecase/authops/login_test.go"
    "internal/usecase/authops/refresh_test.go"
    "internal/usecase/memberops/list_members_test.go"
    "internal/usecase/subops/subscribe_member_test.go"
    "internal/usecase/reservationops/create_reservation_test.go"
    "internal/usecase/paymentops/process_callback_retries_test.go"
    "internal/usecase/paymentops/generate_receipt_test.go"
    "internal/usecase/paymentops/pay_with_saved_card_test.go"
    "internal/usecase/paymentops/refund_payment_test.go"
    "internal/usecase/paymentops/cancel_payment_test.go"
    "internal/usecase/paymentops/list_member_payments_test.go"
    "internal/usecase/paymentops/expire_payments_test.go"
    "internal/usecase/paymentops/verify_payment_test.go"
    "internal/usecase/paymentops/initiate_payment_test.go"
    "internal/usecase/paymentops/handle_callback_test.go"
)

# Function to update imports in a file
update_imports() {
    local file=$1
    echo "  Updating imports in $file..."

    # Check if the file exists
    if [ ! -f "$file" ]; then
        echo "    File not found: $file"
        return 1
    fi

    # Add new imports if not already present
    if ! grep -q "internal/adapters/repository/mocks" "$file"; then
        # Add after the first import statement
        sed -i '' '/^import (/a\
	"library-service/internal/adapters/repository/mocks"
' "$file"
    fi

    if ! grep -q "test/builders" "$file"; then
        sed -i '' '/^import (/a\
	"library-service/test/builders"
' "$file"
    fi

    if ! grep -q "test/helpers" "$file"; then
        sed -i '' '/^import (/a\
	"library-service/test/helpers"
' "$file"
    fi

    # Remove old testutil import if present
    sed -i '' '/test\/testutil/d' "$file"
}

# Function to replace mock types
replace_mock_types() {
    local file=$1
    echo "  Replacing mock types in $file..."

    # Replace mockMemberRepository with mocks.MockMemberRepository
    sed -i '' 's/\*mockMemberRepository/\*mocks.MockMemberRepository/g' "$file"
    sed -i '' 's/new(mockMemberRepository)/new(mocks.MockMemberRepository)/g' "$file"

    # Replace mockBookRepository with mocks.MockBookRepository
    sed -i '' 's/\*mockBookRepository/\*mocks.MockBookRepository/g' "$file"
    sed -i '' 's/new(mockBookRepository)/new(mocks.MockBookRepository)/g' "$file"

    # Replace mockPaymentRepository with mocks.MockPaymentRepository
    sed -i '' 's/\*mockPaymentRepository/\*mocks.MockPaymentRepository/g' "$file"
    sed -i '' 's/new(mockPaymentRepository)/new(mocks.MockPaymentRepository)/g' "$file"

    # Replace mockReservationRepository with mocks.MockReservationRepository
    sed -i '' 's/\*mockReservationRepository/\*mocks.MockReservationRepository/g' "$file"
    sed -i '' 's/new(mockReservationRepository)/new(mocks.MockReservationRepository)/g' "$file"

    # Replace other mock types
    sed -i '' 's/\*mockAuthorRepository/\*mocks.MockAuthorRepository/g' "$file"
    sed -i '' 's/new(mockAuthorRepository)/new(mocks.MockAuthorRepository)/g' "$file"

    sed -i '' 's/\*mockSavedCardRepository/\*mocks.MockSavedCardRepository/g' "$file"
    sed -i '' 's/new(mockSavedCardRepository)/new(mocks.MockSavedCardRepository)/g' "$file"

    sed -i '' 's/\*mockReceiptRepository/\*mocks.MockReceiptRepository/g' "$file"
    sed -i '' 's/new(mockReceiptRepository)/new(mocks.MockReceiptRepository)/g' "$file"

    sed -i '' 's/\*mockCallbackRetryRepository/\*mocks.MockCallbackRetryRepository/g' "$file"
    sed -i '' 's/new(mockCallbackRetryRepository)/new(mocks.MockCallbackRetryRepository)/g' "$file"
}

# Function to replace testutil assertions with helpers
replace_assertions() {
    local file=$1
    echo "  Replacing assertions in $file..."

    # Replace testutil assertions with helpers
    sed -i '' 's/testutil\.AssertEqual/helpers.AssertEqual/g' "$file"
    sed -i '' 's/testutil\.AssertNoError/helpers.AssertNoError/g' "$file"
    sed -i '' 's/testutil\.AssertError/helpers.AssertError/g' "$file"
    sed -i '' 's/testutil\.AssertNil/helpers.AssertNil/g' "$file"
    sed -i '' 's/testutil\.AssertNotNil/helpers.AssertNotNil/g' "$file"
    sed -i '' 's/testutil\.AssertStringContains/helpers.AssertErrorContains/g' "$file"
    sed -i '' 's/testutil\.AssertTrue/helpers.AssertTrue/g' "$file"
    sed -i '' 's/testutil\.AssertFalse/helpers.AssertFalse/g' "$file"

    # Replace context.Background() with helpers.TestContext(t)
    sed -i '' 's/ctx := context\.Background()/ctx := helpers.TestContext(t)/g' "$file"
}

# Process each test file
for file in "${test_files[@]}"; do
    echo "Processing $file..."

    if [ -f "$file" ]; then
        update_imports "$file"
        replace_mock_types "$file"
        replace_assertions "$file"
        echo "  ✅ Updated $file"
    else
        echo "  ⚠️  File not found: $file"
    fi
done

echo ""
echo "✅ Test mock update complete!"
echo ""
echo "Next steps:"
echo "1. Remove old mock definitions from test files"
echo "2. Update test data creation to use builders where appropriate"
echo "3. Run tests to verify everything works: make test"