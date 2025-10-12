#!/bin/bash

# Script to update use cases to use new error patterns
# This updates error handling to use the new fluent error builders

echo "🔄 Updating error patterns in use cases..."

# Find all use case files
USE_CASE_FILES=$(find internal/usecase -name "*.go" -type f ! -name "*_test.go")

# Update common error patterns
for file in $USE_CASE_FILES; do
    echo "Processing: $file"

    # Update ErrBookAlreadyExists pattern
    sed -i '' 's/errors\.ErrBookAlreadyExists\.WithDetails(\([^)]*\))/errors.AlreadyExists("book", \1)/g' "$file"

    # Update ErrMemberAlreadyExists pattern
    sed -i '' 's/errors\.ErrMemberAlreadyExists\.WithDetails(\([^)]*\))/errors.AlreadyExists("member", \1)/g' "$file"

    # Update ErrNotFound patterns
    sed -i '' 's/errors\.ErrNotFound\.WithDetails("entity", "\([^"]*\)")/errors.NotFound("\1")/g' "$file"
    sed -i '' 's/errors\.ErrBookNotFound/errors.NotFound("book")/g' "$file"
    sed -i '' 's/errors\.ErrMemberNotFound/errors.NotFound("member")/g' "$file"
    sed -i '' 's/errors\.ErrAuthorNotFound/errors.NotFound("author")/g' "$file"

    # Update ErrDatabase patterns
    sed -i '' 's/errors\.ErrDatabase\.Wrap(\([^)]*\))/errors.Database("database operation", \1)/g' "$file"
    sed -i '' 's/fmt\.Errorf(".*: %w", \([^)]*\))/errors.Internal("operation failed", \1)/g' "$file"

    # Update ErrUnauthorized patterns
    sed -i '' 's/errors\.ErrUnauthorized/errors.Unauthorized("invalid credentials")/g' "$file"
    sed -i '' 's/errors\.ErrInvalidCredentials/errors.Unauthorized("invalid credentials")/g' "$file"

    # Update ErrValidation patterns
    sed -i '' 's/errors\.ErrValidation\.WithDetails(\([^)]*\))/errors.Validation(\1)/g' "$file"
    sed -i '' 's/errors\.ErrInvalidInput/errors.Validation("input", "invalid format")/g' "$file"

    # Update logger patterns - use UseCaseLogger
    sed -i '' 's/logutil\.UseCaseLogger(ctx, "\([^"]*\)",/logutil.UseCaseLogger(ctx, "\1", "operation")/g' "$file"

    # Update logger usage to include operation
    sed -i '' 's/logger := logutil\.GetLogger()/logger := logutil.FromContext(ctx)/g' "$file"
done

echo ""
echo "🔍 Checking for any remaining old patterns..."

# Check for any remaining old error patterns
echo "Checking for old error patterns..."
grep -r "errors\.Err[A-Z]" internal/usecase/ --include="*.go" | grep -v "_test.go" | head -10

echo ""
echo "✅ Error pattern update complete!"
echo ""
echo "📝 Summary of changes:"
echo "  - Replaced WithDetails() with fluent error builders"
echo "  - Updated error wrapping to use specific error constructors"
echo "  - Standardized logging with UseCaseLogger"
echo "  - Added proper context propagation"

echo ""
echo "🔧 Next steps:"
echo "  1. Review changes with: git diff"
echo "  2. Run tests: make test"
echo "  3. Fix any compilation errors manually"