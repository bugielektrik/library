// Package grpc provides gRPC service implementations for inter-service communication.
//
// This package implements gRPC adapters for high-performance RPC communication
// between microservices. While the primary API is REST/HTTP, gRPC is available
// for internal service-to-service calls requiring lower latency.
//
// Services:
//   - BookService: Book queries and operations
//   - MemberService: Member information and authentication
//   - ReservationService: Reservation status and management
//
// Protocol Buffers:
//   - Proto definitions in api/proto/
//   - Generated Go code via protoc
//   - Versioned API (v1, v2) for backward compatibility
//
// gRPC features used:
//   - Unary RPC: Request/response pattern (similar to REST)
//   - Server streaming: For pagination and bulk operations
//   - Context propagation: Request ID, auth tokens
//   - Metadata: Custom headers, tracing information
//
// Example service implementation:
//
//	type BookServiceServer struct {
//	    getBookUC *bookops.GetBookUseCase
//	}
//
//	func (s *BookServiceServer) GetBook(ctx context.Context, req *pb.GetBookRequest) (*pb.BookResponse, error) {
//	    result, err := s.getBookUC.Execute(ctx, bookops.GetBookRequest{
//	        ID: req.GetId(),
//	    })
//	    if err != nil {
//	        return nil, status.Error(codes.NotFound, err.Error())
//	    }
//	    return toProtoBook(result.Book), nil
//	}
//
// Authentication:
//   - JWT tokens in gRPC metadata (authorization key)
//   - Same authentication middleware as HTTP API
//   - Per-method authorization rules
//
// Error handling:
//   - Domain errors mapped to gRPC status codes
//   - NotFound: codes.NotFound
//   - InvalidArgument: codes.InvalidArgument
//   - Internal: codes.Internal
//   - Unauthenticated: codes.Unauthenticated
//
// Interceptors (middleware):
//   - Authentication: Validate JWT from metadata
//   - Logging: Request/response logging with duration
//   - Recovery: Panic recovery to prevent service crash
//   - Metrics: Request count, latency, error rate
//
// Configuration:
//   - gRPC port (separate from HTTP, default 9090)
//   - TLS certificates for secure communication
//   - Max message size limits
//   - Connection timeout and keepalive
//
// Testing:
//   - Unit tests with mock clients
//   - Integration tests with real gRPC server
//   - Load testing for performance validation
//
// Future use cases:
//   - Inter-service communication in microservices architecture
//   - Mobile app SDK for native performance
//   - Streaming updates for real-time features
package grpc
