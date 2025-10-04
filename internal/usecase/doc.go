// Package usecase contains the application business logic and orchestration.
//
// This is the application layer of Clean Architecture, containing:
//   - Use case implementations (one per business operation)
//   - Use case DTOs for input/output
//   - Orchestration of domain services and repositories
//   - Transaction management
//
// Dependency Rule:
// Use cases depend on domain layer (inward) but never on adapters or infrastructure.
// They use repository interfaces defined in the domain layer.
//
// Use Case Packages:
//   - book: Book-related operations (create, update, delete, list)
//   - member: Member and subscription operations
//   - author: Author management operations
//   - subscription: Subscription-specific business flows
//
// Design Patterns:
//   - One Use Case per File: Single responsibility principle
//   - Constructor Injection: Dependencies injected via constructor
//   - DTO Pattern: Separate DTOs from domain entities
//   - Error Wrapping: Contextual error messages with error chains
//
// Example Use Case Structure:
//
//	// Use case struct with dependencies
//	type CreateBookUseCase struct {
//	    bookRepo    book.Repository
//	    bookService *book.Service
//	    bookCache   book.Cache
//	}
//
//	// Constructor with dependency injection
//	func NewCreateBookUseCase(
//	    repo book.Repository,
//	    service *book.Service,
//	    cache book.Cache,
//	) *CreateBookUseCase {
//	    return &CreateBookUseCase{
//	        bookRepo:    repo,
//	        bookService: service,
//	        bookCache:   cache,
//	    }
//	}
//
//	// Execute method orchestrates the flow
//	func (uc *CreateBookUseCase) Execute(ctx context.Context, input CreateBookInput) (*book.Entity, error) {
//	    // 1. Map DTO to entity
//	    entity := input.ToEntity()
//
//	    // 2. Validate with domain service
//	    if err := uc.bookService.ValidateBook(entity); err != nil {
//	        return nil, fmt.Errorf("validation failed: %w", err)
//	    }
//
//	    // 3. Check business rules
//	    existing, _ := uc.bookRepo.GetByISBN(ctx, entity.ISBN)
//	    if existing != nil {
//	        return nil, errors.New("book already exists")
//	    }
//
//	    // 4. Persist
//	    if err := uc.bookRepo.Create(ctx, entity); err != nil {
//	        return nil, fmt.Errorf("failed to create book: %w", err)
//	    }
//
//	    // 5. Cache
//	    _ = uc.bookCache.Set(ctx, entity)
//
//	    return &entity, nil
//	}
//
// Use cases are thin orchestrators that delegate business logic to domain services,
// ensuring the domain layer remains the source of truth for business rules.
package usecase
