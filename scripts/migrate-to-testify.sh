#!/bin/bash
# Migrate test files from helpers.Assert* to testify

set -e

# Files to migrate
files=(
    "internal/members/service/auth/refresh_test.go"
    "internal/members/service/auth/validate_test.go"
    "internal/members/service/profile/get_member_profile_test.go"
    "internal/members/service/profile/list_members_test.go"
)

for file in "${files[@]}"; do
    echo "Migrating $file..."

    # Add testify imports if not present
    if ! grep -q "github.com/stretchr/testify/assert" "$file"; then
        # Add imports after the first import block
        sed -i '' '/^import (/a\
\	"github.com/stretchr/testify/assert"\
\	"github.com/stretchr/testify/require"
' "$file"
    fi

    # Replace all helpers.Assert* calls with testify equivalents
    sed -i '' 's/helpers\.AssertEqual(/assert.Equal(/g' "$file"
    sed -i '' 's/helpers\.AssertNotEqual(/assert.NotEqual(/g' "$file"
    sed -i '' 's/helpers\.AssertNoError(/require.NoError(/g' "$file"
    sed -i '' 's/helpers\.AssertError(/assert.Error(/g' "$file"
    sed -i '' 's/helpers\.AssertErrorContains(/assert.ErrorContains(/g' "$file"
    sed -i '' 's/helpers\.AssertTrue(\(t, [^)]*\) != "")/assert.NotEmpty(\1)/g' "$file"
    sed -i '' 's/helpers\.AssertTrue(/assert.True(/g' "$file"
    sed -i '' 's/helpers\.AssertFalse(/assert.False(/g' "$file"
    sed -i '' 's/helpers\.AssertNil(/assert.Nil(/g' "$file"
    sed -i '' 's/helpers\.AssertNotNil(/assert.NotNil(/g' "$file"
    sed -i '' 's/helpers\.AssertContains(/assert.Contains(/g' "$file"
    sed -i '' 's/helpers\.AssertLen(/assert.Len(/g' "$file"

    echo "✓ $file migrated"
done

echo ""
echo "✓ All files migrated to testify!"
echo ""
echo "Running goimports to organize imports..."
goimports -w "${files[@]}"

echo "✓ Done!"
