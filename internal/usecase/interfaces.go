package usecase

import "context"

// UseCase represents a single business use case
type UseCase[TRequest, TResponse any] interface {
	Execute(ctx context.Context, req TRequest) (TResponse, error)
}

// UseCaseWithoutResponse represents a use case that doesn't return data
type UseCaseWithoutResponse[TRequest any] interface {
	Execute(ctx context.Context, req TRequest) error
}

// QueryUseCase represents a read-only use case
type QueryUseCase[TRequest, TResponse any] interface {
	Execute(ctx context.Context, req TRequest) (TResponse, error)
}

// CommandUseCase represents a write operation use case
type CommandUseCase[TRequest, TResponse any] interface {
	Execute(ctx context.Context, req TRequest) (TResponse, error)
}
