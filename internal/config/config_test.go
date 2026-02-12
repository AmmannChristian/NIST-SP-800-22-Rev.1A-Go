package config

import (
	"testing"
)

func TestLoadWithEnvOverrides(t *testing.T) {
	t.Setenv("GRPC_PORT", "5000")
	t.Setenv("METRICS_PORT", "6000")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_ISSUER", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "nist-api")
	t.Setenv("AUTH_JWKS_URL", "https://issuer.example.com/jwks.json")
	t.Setenv("AUTH_TOKEN_TYPE", "jwt")
	t.Setenv("TLS_ENABLED", "true")
	t.Setenv("TLS_CERT_FILE", "/tmp/cert.pem")
	t.Setenv("TLS_KEY_FILE", "/tmp/key.pem")
	t.Setenv("TLS_CA_FILE", "/tmp/ca.pem")
	t.Setenv("TLS_CLIENT_AUTH", "requireandverify")
	t.Setenv("TLS_MIN_VERSION", "1.3")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.GRPCPort != 5000 || cfg.MetricsPort != 6000 {
		t.Fatalf("unexpected ports: %+v", cfg)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("unexpected log level: %s", cfg.LogLevel)
	}
	if !cfg.AuthEnabled {
		t.Fatalf("expected AuthEnabled to be true")
	}
	if cfg.AuthIssuer != "https://issuer.example.com" {
		t.Fatalf("unexpected issuer: %s", cfg.AuthIssuer)
	}
	if cfg.AuthAudience != "nist-api" {
		t.Fatalf("unexpected audience: %s", cfg.AuthAudience)
	}
	if cfg.AuthJWKSURL != "https://issuer.example.com/jwks.json" {
		t.Fatalf("unexpected JWKS URL: %s", cfg.AuthJWKSURL)
	}
	if cfg.AuthTokenType != "jwt" {
		t.Fatalf("unexpected auth token type: %s", cfg.AuthTokenType)
	}
	if !cfg.TLSEnabled {
		t.Fatalf("expected TLSEnabled to be true")
	}
	if cfg.TLSCertFile != "/tmp/cert.pem" || cfg.TLSKeyFile != "/tmp/key.pem" || cfg.TLSCAFile != "/tmp/ca.pem" {
		t.Fatalf("unexpected TLS file config: %+v", cfg)
	}
	if cfg.TLSClientAuth != "requireandverify" {
		t.Fatalf("unexpected TLS client auth: %s", cfg.TLSClientAuth)
	}
	if cfg.TLSMinVersion != "1.3" {
		t.Fatalf("unexpected TLS min version: %s", cfg.TLSMinVersion)
	}
}

func TestLoadWithOpaqueAuthEnvOverrides(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_ISSUER", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "nist-api")
	t.Setenv("AUTH_TOKEN_TYPE", "opaque")
	t.Setenv("AUTH_INTROSPECTION_URL", "https://issuer.example.com/oauth2/introspect")
	t.Setenv("AUTH_INTROSPECTION_CLIENT_ID", "svc-client")
	t.Setenv("AUTH_INTROSPECTION_CLIENT_SECRET", "svc-secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AuthTokenType != "opaque" {
		t.Fatalf("expected opaque auth token type, got %s", cfg.AuthTokenType)
	}
	if cfg.AuthIntrospectionURL != "https://issuer.example.com/oauth2/introspect" {
		t.Fatalf("unexpected introspection URL: %s", cfg.AuthIntrospectionURL)
	}
	if cfg.AuthIntrospectionClientID != "svc-client" {
		t.Fatalf("unexpected introspection client id: %s", cfg.AuthIntrospectionClientID)
	}
	if cfg.AuthIntrospectionClientSecret != "svc-secret" {
		t.Fatalf("unexpected introspection client secret: %s", cfg.AuthIntrospectionClientSecret)
	}
}

func TestValidateFailures(t *testing.T) {
	tests := []struct {
		name string
		cfg  Config
	}{
		{"bad grpc port", Config{GRPCPort: 0, MetricsPort: 9000, LogLevel: "info"}},
		{"bad metrics port", Config{GRPCPort: 9000, MetricsPort: 70000, LogLevel: "info"}},
		{"bad log level", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "verbose"}},
		{"auth enabled missing issuer", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthAudience: "api"}},
		{"auth enabled missing audience", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com"}},
		{"auth enabled invalid token type", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "paseto"}},
		{"auth opaque missing introspection url", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionClientID: "svc-client", AuthIntrospectionClientSecret: "svc-secret"}},
		{"auth opaque missing introspection client id", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionURL: "https://issuer.example.com/oauth2/introspect", AuthIntrospectionClientSecret: "svc-secret"}},
		{"auth opaque missing introspection client secret", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionURL: "https://issuer.example.com/oauth2/introspect", AuthIntrospectionClientID: "svc-client"}},
		{"tls enabled missing cert", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", TLSEnabled: true, TLSKeyFile: "/tmp/key.pem"}},
		{"tls enabled missing key", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", TLSEnabled: true, TLSCertFile: "/tmp/cert.pem"}},
		{"tls enabled invalid client auth", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", TLSEnabled: true, TLSCertFile: "/tmp/cert.pem", TLSKeyFile: "/tmp/key.pem", TLSClientAuth: "invalid"}},
		{"tls enabled invalid min version", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", TLSEnabled: true, TLSCertFile: "/tmp/cert.pem", TLSKeyFile: "/tmp/key.pem", TLSMinVersion: "1.1"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Validate(); err == nil {
				t.Fatalf("expected error for %s", tt.name)
			}
		})
	}

	// getEnvInt falls back on parse error
	t.Setenv("SOME_INT", "notanint")
	if v := getEnvInt("SOME_INT", 42); v != 42 {
		t.Fatalf("expected default on parse error, got %d", v)
	}
}

func TestValidateOpaqueAuthSuccess(t *testing.T) {
	cfg := Config{
		GRPCPort:                      9000,
		MetricsPort:                   9001,
		LogLevel:                      "info",
		AuthEnabled:                   true,
		AuthIssuer:                    "https://issuer.example.com",
		AuthAudience:                  "api",
		AuthTokenType:                 "opaque",
		AuthIntrospectionURL:          "https://issuer.example.com/oauth2/introspect",
		AuthIntrospectionClientID:     "svc-client",
		AuthIntrospectionClientSecret: "svc-secret",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid opaque auth config, got error: %v", err)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Clear any environment variables
	for _, key := range []string{
		"GRPC_PORT", "METRICS_PORT", "LOG_LEVEL",
		"AUTH_ENABLED", "AUTH_ISSUER", "AUTH_AUDIENCE", "AUTH_JWKS_URL", "AUTH_TOKEN_TYPE",
		"AUTH_INTROSPECTION_URL", "AUTH_INTROSPECTION_CLIENT_ID", "AUTH_INTROSPECTION_CLIENT_SECRET",
		"TLS_ENABLED", "TLS_CERT_FILE", "TLS_KEY_FILE", "TLS_CA_FILE", "TLS_CLIENT_AUTH", "TLS_MIN_VERSION",
	} {
		t.Setenv(key, "")
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	// Check default values
	if cfg.GRPCPort != 9090 {
		t.Errorf("expected default GRPCPort=9090, got %d", cfg.GRPCPort)
	}
	if cfg.MetricsPort != 9091 {
		t.Errorf("expected default MetricsPort=9091, got %d", cfg.MetricsPort)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default LogLevel=info, got %s", cfg.LogLevel)
	}
	if cfg.AuthEnabled {
		t.Errorf("expected AuthEnabled to be false by default")
	}
	if cfg.AuthIssuer != "" || cfg.AuthAudience != "" || cfg.AuthJWKSURL != "" || cfg.AuthIntrospectionURL != "" || cfg.AuthIntrospectionClientID != "" || cfg.AuthIntrospectionClientSecret != "" {
		t.Errorf("expected auth config defaults to be empty, got %+v", cfg)
	}
	if cfg.AuthTokenType != "jwt" {
		t.Errorf("expected AuthTokenType to default to 'jwt', got %s", cfg.AuthTokenType)
	}
	if cfg.TLSEnabled {
		t.Errorf("expected TLSEnabled to be false by default")
	}
	if cfg.TLSCertFile != "" || cfg.TLSKeyFile != "" || cfg.TLSCAFile != "" {
		t.Errorf("expected TLS file settings to be empty by default, got %+v", cfg)
	}
	if cfg.TLSClientAuth != "none" {
		t.Errorf("expected TLSClientAuth to default to 'none', got %s", cfg.TLSClientAuth)
	}
	if cfg.TLSMinVersion != "1.2" {
		t.Errorf("expected TLSMinVersion to default to '1.2', got %s", cfg.TLSMinVersion)
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	t.Setenv("GRPC_PORT", "0")
	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}
