#!/bin/bash

# Fix remaining payment error patterns

echo "🔄 Fixing remaining payment error patterns..."

# Fix specific payment errors
find internal/usecase/paymentops -name "*.go" -type f ! -name "*_test.go" | while read file; do
    echo "Processing: $file"

    # Fix ErrNotFound patterns with specific entities
    sed -i '' 's/errors\.ErrNotFound\.WithDetails("card_id", \([^)]*\))/errors.NotFoundWithID("card", \1)/g' "$file"
    sed -i '' 's/errors\.ErrNotFound\.WithDetails("receipt_id", \([^)]*\))/errors.NotFoundWithID("receipt", \1)/g' "$file"
    sed -i '' 's/errors\.ErrNotFound\.WithDetails("payment_id", \([^)]*\))/errors.NotFoundWithID("payment", \1)/g' "$file"
    sed -i '' 's/errors\.ErrNotFound\.WithDetails("member_id", \([^)]*\))/errors.NotFoundWithID("member", \1)/g' "$file"

    # Fix payment-specific errors
    sed -i '' 's/errors\.ErrPaymentNotFound/errors.NotFound("payment")/g' "$file"
    sed -i '' 's/errors\.ErrPaymentGateway\.Wrap(\([^)]*\))/errors.External("payment provider", \1)/g' "$file"
    sed -i '' 's/errors\.ErrPaymentGateway/errors.External("payment provider", nil)/g' "$file"
    sed -i '' 's/errors\.ErrDatabase\./errors.Database("database operation", nil)./g' "$file"

    # Fix validation errors
    sed -i '' 's/errors\.ErrValidation\./errors.NewError(errors.CodeValidation)./g' "$file"
done

echo ""
echo "✅ Payment error patterns fixed!"