#!/bin/bash
# Script to convert handler to use the generic wrapper pattern

echo "Converting handlers to use generic wrapper pattern..."

# Function to create optimized handler template
create_optimized_handler() {
    local package=$1
    local handler_name=$2
    local file_path=$3

    cat > "$file_path" << 'EOF'
package PACKAGE_NAME

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/adapters/http/handlers"
	"library-service/internal/usecase"
	"library-service/pkg/httputil"
)

// OptimizedHANDLER_NAMEHandler is the optimized handler using generic wrappers
type OptimizedHANDLER_NAMEHandler struct {
	useCases  *usecase.Container
	validator *handlers.ValidatorAdapter
}

// NewOptimizedHANDLER_NAMEHandler creates a new optimized handler
func NewOptimizedHANDLER_NAMEHandler(
	useCases *usecase.Container,
	validator *middleware.Validator,
) *OptimizedHANDLER_NAMEHandler {
	return &OptimizedHANDLER_NAMEHandler{
		useCases:  useCases,
		validator: handlers.NewValidatorAdapter(validator),
	}
}

// Routes returns the routes
func (h *OptimizedHANDLER_NAMEHandler) Routes() chi.Router {
	r := chi.NewRouter()
	// TODO: Add routes here
	return r
}

// Example handler method using wrapper
func (h *OptimizedHANDLER_NAMEHandler) ExampleOperation() http.HandlerFunc {
	return httputil.CreateHandler(
		// TODO: Replace with actual use case
		h.useCases.DOMAIN.Operation.Execute,
		h.validator.CreateValidator[RequestType](),
		"handler_name", "operation",
		httputil.WrapperOptions{RequireAuth: true},
	)
}
EOF

    # Replace placeholders
    sed -i '' "s/PACKAGE_NAME/$package/g" "$file_path"
    sed -i '' "s/HANDLER_NAME/$handler_name/g" "$file_path"
}

# Handler directories to convert
handler=(
    "book:Book"
    "payment:Payment"
    "reservation:Reservation"
    "savedcard:SavedCard"
)

for handler_info in "${handler[@]}"; do
    IFS=':' read -r package name <<< "$handler_info"

    output_file="internal/adapters/http/handlers/$package/handler_optimized.go"

    if [ ! -f "$output_file" ]; then
        echo "Creating optimized handler for $package..."
        create_optimized_handler "$package" "$name" "$output_file"
        echo "  ✅ Created $output_file"
    else
        echo "  ⚠️  $output_file already exists, skipping"
    fi
done

echo ""
echo "✅ Handler conversion templates created!"
echo ""
echo "Next steps:"
echo "1. Review each generated handler_optimized.go file"
echo "2. Copy handler logic from original handler.go"
echo "3. Convert each method to use the wrapper pattern"
echo "4. Test the converted handlers"
echo "5. Replace original handlers with optimized versions"