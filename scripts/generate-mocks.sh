#!/bin/bash
# Generate centralized mocks for all repository interfaces

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Generating repository mocks with mockery...${NC}"

# Create mocks directory if it doesn't exist
mkdir -p internal/adapters/repository/mocks

# Change to project root
cd "$(dirname "$0")/.."

# Generate mocks for all repository interfaces
echo "→ Generating repository mocks..."
~/go/bin/mockery --dir=internal/domain --name=Repository --output=internal/adapters/repository/mocks --outpkg=mocks --all

# Generate mocks for cache interfaces if they exist
echo "→ Generating cache mocks..."
~/go/bin/mockery --dir=internal/domain --name=Cache --output=internal/adapters/repository/mocks --outpkg=mocks --all 2>/dev/null || echo "  (no cache interfaces found)"

# Generate mocks for SavedCardRepository interface
echo "→ Generating SavedCard repository mock..."
~/go/bin/mockery --dir=internal/domain/payment --name=SavedCardRepository --output=internal/adapters/repository/mocks --outpkg=mocks

# Generate mocks for ReceiptRepository interface
echo "→ Generating Receipt repository mock..."
~/go/bin/mockery --dir=internal/domain/payment --name=ReceiptRepository --output=internal/adapters/repository/mocks --outpkg=mocks

# Generate mocks for CallbackRetryRepository interface
echo "→ Generating CallbackRetry repository mock..."
~/go/bin/mockery --dir=internal/domain/payment --name=CallbackRetryRepository --output=internal/adapters/repository/mocks --outpkg=mocks

echo -e "${GREEN}✓ All mocks generated successfully!${NC}"
echo ""
echo "Mocks are located in: internal/adapters/repository/mocks/"
echo ""
echo "To use in tests:"
echo '  import ('
echo '      "github.com/stretchr/testify/mock"'
echo '      "library-service/internal/adapters/repository/mocks"'
echo '  )'
echo '  mockRepo := mocks.NewMockRepository(t)'
echo '  mockRepo.On("Get", mock.Anything, "123").Return(book, nil)'