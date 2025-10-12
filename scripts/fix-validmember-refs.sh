#!/bin/bash
# Script to fix validMember references to use builders

echo "Fixing validMember references in test files..."

# Replace validMember() calls with builders.Member().Build()
sed -i '' 's/validMember()/builders.Member().Build()/g' internal/usecase/memberops/list_members_test.go

# Also add the builders import if not present
if ! grep -q "test/builders" internal/usecase/memberops/list_members_test.go; then
    sed -i '' '/^import (/a\
	"library-service/test/builders"
' internal/usecase/memberops/list_members_test.go
fi

echo "✅ Fixed validMember references"