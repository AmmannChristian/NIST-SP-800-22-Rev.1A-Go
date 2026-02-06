// Package middleware provides gRPC server interceptors for the NIST SP 800-22
// service, including request-ID generation for log correlation.
package middleware

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// contextKey is an unexported type used as a key for context.WithValue to
// prevent collisions with keys defined in other packages.
type contextKey string

// RequestIDKey is the context key under which the request ID is stored.
const RequestIDKey contextKey = "request_id"

// UnaryRequestIDInterceptor returns a gRPC unary server interceptor that
// generates a UUID v4 request ID for each incoming call. The ID is stored in the
// context for internal use and propagated to the client via the "x-request-id"
// response header metadata.
func UnaryRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Generate unique request ID
		requestID := uuid.New().String()

		// Add to context for internal use
		ctx = context.WithValue(ctx, RequestIDKey, requestID)

		// Add to outgoing metadata for clients to receive
		md := metadata.Pairs("x-request-id", requestID)
		if err := grpc.SetHeader(ctx, md); err != nil {
			// Continue even if header setting fails
			// This is not critical for request processing
		}

		return handler(ctx, req)
	}
}

// GetRequestID extracts the request ID previously set by UnaryRequestIDInterceptor.
// It returns an empty string if no request ID is present in the context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
