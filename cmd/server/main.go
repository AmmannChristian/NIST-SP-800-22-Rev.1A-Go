// Package main provides the entry point for the NIST SP 800-22 Rev 1a
// Statistical Test Service. It bootstraps a gRPC server with optional TLS
// and OIDC authentication, exposes a Prometheus metrics endpoint with pprof
// profiling, and handles graceful shutdown on SIGTERM and SIGINT.
package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/AmmannChristian/go-authx/grpcserver"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/AmmannChristian/nist-sp800-22-rev1a/internal/config"
	"github.com/AmmannChristian/nist-sp800-22-rev1a/internal/middleware"
	"github.com/AmmannChristian/nist-sp800-22-rev1a/internal/service"
	pb "github.com/AmmannChristian/nist-sp800-22-rev1a/pkg/pb"
)

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Application failed")
	}
}

// run initialises the service components (configuration, logging, metrics server,
// gRPC server with interceptors) and blocks until a termination signal is received
// or the provided context is cancelled.
func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	setupLogging(cfg.LogLevel)

	log.Info().
		Int("grpc_port", cfg.GRPCPort).
		Int("metrics_port", cfg.MetricsPort).
		Str("log_level", cfg.LogLevel).
		Bool("auth_enabled", cfg.AuthEnabled).
		Msg("Starting NIST Statistical Test Service")

	metricsLn, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.MetricsPort))
	if err != nil {
		return fmt.Errorf("failed to create metrics listener: %w", err)
	}
	metricsSrv := startMetricsServer(metricsLn)
	defer metricsSrv.Close()

	grpcLn, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to create gRPC listener: %w", err)
	}

	unaryInterceptors, err := buildUnaryInterceptors(cfg)
	if err != nil {
		return fmt.Errorf("failed to configure gRPC server: %w", err)
	}

	grpcServer, err := runGRPCServer(cfg, unaryInterceptors)
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		select {
		case <-sigChan:
			log.Info().Msg("Shutting down gracefully...")
		case <-ctx.Done():
		}

		grpcServer.GracefulStop()
		cancel()
	}()

	log.Info().
		Int("port", cfg.GRPCPort).
		Msg("gRPC server listening")

	if err := grpcServer.Serve(grpcLn); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// setupLogging configures the global zerolog logger with a human-readable console
// writer and sets the log level according to the provided string. Unrecognised
// level strings default to "info".
func setupLogging(level string) {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// startMetricsServer creates and starts an HTTP server on the given listener,
// exposing Prometheus metrics at /metrics and a JSON health endpoint at /health.
// The server runs in a background goroutine; the caller is responsible for
// closing the returned *http.Server when shutting down.
func startMetricsServer(ln net.Listener) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{
			"status":  "healthy",
			"version": service.Version,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Error().Err(err).Msg("failed to write health response")
			http.Error(w, "failed to encode health response", http.StatusInternalServerError)
		}
	})

	log.Info().
		Str("addr", ln.Addr().String()).
		Msg("Metrics server listening (with pprof at /debug/pprof/)")

	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Metrics server failed")
		}
	}()

	return srv
}

// runGRPCServer creates a gRPC server with the given interceptors, registers the
// NIST SP 800-22 test service, the gRPC health check service, and server reflection.
func runGRPCServer(cfg *config.Config, unaryInterceptors []grpc.UnaryServerInterceptor) (*grpc.Server, error) {
	serverOpts, err := buildGRPCServerOptions(cfg, unaryInterceptors)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer(serverOpts...)

	nistServer := service.NewServer()
	pb.RegisterSp80022TestServiceServer(grpcServer, nistServer)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	return grpcServer, nil
}

// buildUnaryInterceptors assembles the chain of gRPC unary interceptors. The chain
// always includes request-ID injection and request logging. When authentication is
// enabled, an OIDC token validator interceptor is appended that exempts the
// gRPC health check methods from authentication. Validation supports JWT (JWKS)
// and opaque tokens (OAuth2 introspection).
func buildUnaryInterceptors(cfg *config.Config) ([]grpc.UnaryServerInterceptor, error) {
	interceptors := []grpc.UnaryServerInterceptor{
		middleware.UnaryRequestIDInterceptor(),
		loggingInterceptor,
	}

	if !cfg.AuthEnabled {
		return interceptors, nil
	}

	validatorBuilder := grpcserver.NewValidatorBuilder(cfg.AuthIssuer, cfg.AuthAudience)
	if cfg.AuthTokenType == "opaque" {
		validatorBuilder = validatorBuilder.WithOpaqueTokenIntrospection(
			cfg.AuthIntrospectionURL,
			cfg.AuthIntrospectionClientID,
			cfg.AuthIntrospectionClientSecret,
		)
	} else if cfg.AuthJWKSURL != "" {
		validatorBuilder = validatorBuilder.WithJWKSURL(cfg.AuthJWKSURL)
	}

	validator, err := validatorBuilder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build auth validator: %w", err)
	}

	log.Info().
		Str("token_type", cfg.AuthTokenType).
		Str("issuer", cfg.AuthIssuer).
		Str("audience", cfg.AuthAudience).
		Str("jwks_url", cfg.AuthJWKSURL).
		Str("introspection_url", cfg.AuthIntrospectionURL).
		Msg("gRPC authentication enabled")

	authInterceptor := grpcserver.UnaryServerInterceptor(
		validator,
		grpcserver.WithExemptMethods(
			"/grpc.health.v1.Health/Check",
			"/grpc.health.v1.Health/Watch",
		),
	)

	return append(interceptors, authInterceptor), nil
}

// buildGRPCServerOptions assembles gRPC server options from configuration. When TLS
// is enabled, it constructs TLS credentials from the configured certificate, key,
// optional CA file, client authentication mode, and minimum TLS version.
func buildGRPCServerOptions(cfg *config.Config, unaryInterceptors []grpc.UnaryServerInterceptor) ([]grpc.ServerOption, error) {
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
	}

	if !cfg.TLSEnabled {
		return opts, nil
	}

	clientAuth, err := cfg.TLSClientAuthType()
	if err != nil {
		return nil, fmt.Errorf("invalid TLS client auth setting: %w", err)
	}

	minVersion, err := cfg.TLSMinVersionValue()
	if err != nil {
		return nil, fmt.Errorf("invalid TLS min version: %w", err)
	}

	tlsConfig := &grpcserver.TLSConfig{
		CertFile:   cfg.TLSCertFile,
		KeyFile:    cfg.TLSKeyFile,
		CAFile:     cfg.TLSCAFile,
		ClientAuth: clientAuth,
		MinVersion: minVersion,
	}

	tlsOpt, err := grpcserver.ServerOption(tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to configure TLS: %w", err)
	}

	log.Info().
		Bool("tls_enabled", true).
		Str("cert_file", cfg.TLSCertFile).
		Str("key_file", cfg.TLSKeyFile).
		Str("ca_file", cfg.TLSCAFile).
		Str("client_auth", cfg.TLSClientAuth).
		Str("min_version", tlsVersionString(minVersion)).
		Msg("gRPC TLS enabled")

	return append(opts, tlsOpt), nil
}

// tlsVersionString returns a human-readable label for the given TLS version constant.
func tlsVersionString(version uint16) string {
	switch version {
	case tls.VersionTLS13:
		return "1.3"
	case tls.VersionTLS12:
		return "1.2"
	default:
		return fmt.Sprintf("0x%x", version)
	}
}

// loggingInterceptor is a gRPC unary server interceptor that logs every request
// with its associated request ID, method name, duration, and outcome.
func loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()

	requestID := middleware.GetRequestID(ctx)

	resp, err := handler(ctx, req)

	duration := time.Since(start)

	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("method", info.FullMethod).
			Dur("duration", duration).
			Msg("gRPC request failed")
	} else {
		log.Debug().
			Str("request_id", requestID).
			Str("method", info.FullMethod).
			Dur("duration", duration).
			Msg("gRPC request completed")
	}

	return resp, err
}
