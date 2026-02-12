package config

import (
	"os"
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
	t.Setenv("AUTHZ_REQUIRED_ROLES", "NIST_ROLE, entropy-admin ")
	t.Setenv("AUTHZ_REQUIRED_SCOPES", "openid, profile")
	t.Setenv("AUTHZ_ROLE_MATCH_MODE", "all")
	t.Setenv("AUTHZ_SCOPE_MATCH_MODE", "any")
	t.Setenv("AUTHZ_ROLE_CLAIM_PATHS", "roles,urn:zitadel:iam:org:project:roles")
	t.Setenv("AUTHZ_SCOPE_CLAIM_PATHS", "scope,scp")
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
	if len(cfg.AuthzRequiredRoles) != 2 || cfg.AuthzRequiredRoles[0] != "NIST_ROLE" || cfg.AuthzRequiredRoles[1] != "entropy-admin" {
		t.Fatalf("unexpected authz required roles: %#v", cfg.AuthzRequiredRoles)
	}
	if len(cfg.AuthzRequiredScopes) != 2 || cfg.AuthzRequiredScopes[0] != "openid" || cfg.AuthzRequiredScopes[1] != "profile" {
		t.Fatalf("unexpected authz required scopes: %#v", cfg.AuthzRequiredScopes)
	}
	if cfg.AuthzRoleMatchMode != "all" {
		t.Fatalf("unexpected authz role match mode: %s", cfg.AuthzRoleMatchMode)
	}
	if cfg.AuthzScopeMatchMode != "any" {
		t.Fatalf("unexpected authz scope match mode: %s", cfg.AuthzScopeMatchMode)
	}
	if len(cfg.AuthzRoleClaimPaths) != 2 || cfg.AuthzRoleClaimPaths[0] != "roles" || cfg.AuthzRoleClaimPaths[1] != "urn:zitadel:iam:org:project:roles" {
		t.Fatalf("unexpected authz role claim paths: %#v", cfg.AuthzRoleClaimPaths)
	}
	if len(cfg.AuthzScopeClaimPaths) != 2 || cfg.AuthzScopeClaimPaths[0] != "scope" || cfg.AuthzScopeClaimPaths[1] != "scp" {
		t.Fatalf("unexpected authz scope claim paths: %#v", cfg.AuthzScopeClaimPaths)
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
	t.Setenv("AUTH_INTROSPECTION_AUTH_METHOD", "client_secret_basic")
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
	if cfg.AuthIntrospectionAuthMethod != "client_secret_basic" {
		t.Fatalf("unexpected introspection auth method: %s", cfg.AuthIntrospectionAuthMethod)
	}
	if cfg.AuthIntrospectionClientID != "svc-client" {
		t.Fatalf("unexpected introspection client id: %s", cfg.AuthIntrospectionClientID)
	}
	if cfg.AuthIntrospectionClientSecret != "svc-secret" {
		t.Fatalf("unexpected introspection client secret: %s", cfg.AuthIntrospectionClientSecret)
	}
}

func TestLoadWithOpaquePrivateKeyJWTEnvOverrides(t *testing.T) {
	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_ISSUER", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "nist-api")
	t.Setenv("AUTH_TOKEN_TYPE", "opaque")
	t.Setenv("AUTH_INTROSPECTION_URL", "https://issuer.example.com/oauth2/introspect")
	t.Setenv("AUTH_INTROSPECTION_AUTH_METHOD", "private_key_jwt")
	t.Setenv("AUTH_INTROSPECTION_PRIVATE_KEY", `{"keyId":"kid-1","key":"-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----","clientId":"svc-client"}`)
	t.Setenv("AUTH_INTROSPECTION_PRIVATE_KEY_JWT_ALG", "rs256")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.AuthIntrospectionAuthMethod != "private_key_jwt" {
		t.Fatalf("unexpected introspection auth method: %s", cfg.AuthIntrospectionAuthMethod)
	}
	if cfg.AuthIntrospectionPrivateKey != `{"keyId":"kid-1","key":"-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----","clientId":"svc-client"}` {
		t.Fatalf("unexpected private key value: %s", cfg.AuthIntrospectionPrivateKey)
	}
	if cfg.AuthIntrospectionPrivateKeyJWTAlgorithm != "RS256" {
		t.Fatalf("unexpected private key jwt alg value: %s", cfg.AuthIntrospectionPrivateKeyJWTAlgorithm)
	}
}

func TestLoadWithOpaquePrivateKeyJWTFromFile(t *testing.T) {
	privateKeyFile, err := os.CreateTemp(t.TempDir(), "zitadel-key-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_, err = privateKeyFile.WriteString(`{"keyId":"kid-from-file","key":"-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----","clientId":"svc-client"}`)
	if err != nil {
		t.Fatalf("failed to write temp key file: %v", err)
	}
	if err := privateKeyFile.Close(); err != nil {
		t.Fatalf("failed to close temp key file: %v", err)
	}

	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_ISSUER", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "nist-api")
	t.Setenv("AUTH_TOKEN_TYPE", "opaque")
	t.Setenv("AUTH_INTROSPECTION_URL", "https://issuer.example.com/oauth2/introspect")
	t.Setenv("AUTH_INTROSPECTION_AUTH_METHOD", "private_key_jwt")
	t.Setenv("AUTH_INTROSPECTION_PRIVATE_KEY_FILE", privateKeyFile.Name())

	cfg, loadErr := Load()
	if loadErr != nil {
		t.Fatalf("Load() returned error: %v", loadErr)
	}

	if cfg.AuthIntrospectionAuthMethod != "private_key_jwt" {
		t.Fatalf("unexpected introspection auth method: %s", cfg.AuthIntrospectionAuthMethod)
	}
	if cfg.AuthIntrospectionPrivateKey != `{"keyId":"kid-from-file","key":"-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----","clientId":"svc-client"}` {
		t.Fatalf("unexpected private key value: %s", cfg.AuthIntrospectionPrivateKey)
	}
	if cfg.AuthIntrospectionPrivateKeyFile != privateKeyFile.Name() {
		t.Fatalf("unexpected private key file path: %s", cfg.AuthIntrospectionPrivateKeyFile)
	}
}

func TestLoadWithOpaquePrivateKeyJWTFromEmptyFileFails(t *testing.T) {
	privateKeyFile, err := os.CreateTemp(t.TempDir(), "empty-zitadel-key-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if err := privateKeyFile.Close(); err != nil {
		t.Fatalf("failed to close temp key file: %v", err)
	}

	t.Setenv("AUTH_ENABLED", "true")
	t.Setenv("AUTH_ISSUER", "https://issuer.example.com")
	t.Setenv("AUTH_AUDIENCE", "nist-api")
	t.Setenv("AUTH_TOKEN_TYPE", "opaque")
	t.Setenv("AUTH_INTROSPECTION_URL", "https://issuer.example.com/oauth2/introspect")
	t.Setenv("AUTH_INTROSPECTION_AUTH_METHOD", "private_key_jwt")
	t.Setenv("AUTH_INTROSPECTION_PRIVATE_KEY_FILE", privateKeyFile.Name())

	if _, loadErr := Load(); loadErr == nil {
		t.Fatal("expected error for empty AUTH_INTROSPECTION_PRIVATE_KEY_FILE")
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
		{"auth opaque missing introspection url", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionAuthMethod: "client_secret_basic", AuthIntrospectionClientID: "svc-client", AuthIntrospectionClientSecret: "svc-secret"}},
		{"auth opaque missing introspection client id", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionURL: "https://issuer.example.com/oauth2/introspect", AuthIntrospectionAuthMethod: "client_secret_basic", AuthIntrospectionClientSecret: "svc-secret"}},
		{"auth opaque missing introspection client secret", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionURL: "https://issuer.example.com/oauth2/introspect", AuthIntrospectionAuthMethod: "client_secret_basic", AuthIntrospectionClientID: "svc-client"}},
		{"auth opaque invalid introspection auth method", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionURL: "https://issuer.example.com/oauth2/introspect", AuthIntrospectionAuthMethod: "mtls"}},
		{"auth opaque private key jwt missing private key", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionURL: "https://issuer.example.com/oauth2/introspect", AuthIntrospectionAuthMethod: "private_key_jwt"}},
		{"auth opaque private key jwt both inline and file set", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionURL: "https://issuer.example.com/oauth2/introspect", AuthIntrospectionAuthMethod: "private_key_jwt", AuthIntrospectionPrivateKey: "PEM", AuthIntrospectionPrivateKeyFile: "/tmp/key.json"}},
		{"auth opaque private key jwt invalid algorithm", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthEnabled: true, AuthIssuer: "https://issuer.example.com", AuthAudience: "api", AuthTokenType: "opaque", AuthIntrospectionURL: "https://issuer.example.com/oauth2/introspect", AuthIntrospectionAuthMethod: "private_key_jwt", AuthIntrospectionPrivateKey: "PEM", AuthIntrospectionPrivateKeyJWTAlgorithm: "PS256"}},
		{"authz invalid role match mode", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthzRoleMatchMode: "one"}},
		{"authz invalid scope match mode", Config{GRPCPort: 9000, MetricsPort: 9001, LogLevel: "info", AuthzScopeMatchMode: "one"}},
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
		AuthIntrospectionAuthMethod:   "client_secret_basic",
		AuthIntrospectionClientID:     "svc-client",
		AuthIntrospectionClientSecret: "svc-secret",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid opaque auth config, got error: %v", err)
	}
}

func TestValidateOpaquePrivateKeyJWTAuthSuccess(t *testing.T) {
	cfg := Config{
		GRPCPort:                                9000,
		MetricsPort:                             9001,
		LogLevel:                                "info",
		AuthEnabled:                             true,
		AuthIssuer:                              "https://issuer.example.com",
		AuthAudience:                            "api",
		AuthTokenType:                           "opaque",
		AuthIntrospectionURL:                    "https://issuer.example.com/oauth2/introspect",
		AuthIntrospectionAuthMethod:             "private_key_jwt",
		AuthIntrospectionPrivateKey:             "PEM",
		AuthIntrospectionPrivateKeyJWTAlgorithm: "es256",
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid opaque private_key_jwt config, got error: %v", err)
	}
	if cfg.AuthIntrospectionPrivateKeyJWTAlgorithm != "ES256" {
		t.Fatalf("expected algorithm to normalize to ES256, got %s", cfg.AuthIntrospectionPrivateKeyJWTAlgorithm)
	}
}

func TestLoadDefaults(t *testing.T) {
	// Clear any environment variables
	for _, key := range []string{
		"GRPC_PORT", "METRICS_PORT", "LOG_LEVEL",
		"AUTH_ENABLED", "AUTH_ISSUER", "AUTH_AUDIENCE", "AUTH_JWKS_URL", "AUTH_TOKEN_TYPE",
		"AUTH_INTROSPECTION_URL", "AUTH_INTROSPECTION_AUTH_METHOD", "AUTH_INTROSPECTION_CLIENT_ID", "AUTH_INTROSPECTION_CLIENT_SECRET",
		"AUTH_INTROSPECTION_PRIVATE_KEY", "AUTH_INTROSPECTION_PRIVATE_KEY_FILE",
		"AUTH_INTROSPECTION_PRIVATE_KEY_JWT_KID", "AUTH_INTROSPECTION_PRIVATE_KEY_JWT_ALG",
		"AUTHZ_REQUIRED_ROLES", "AUTHZ_REQUIRED_SCOPES", "AUTHZ_ROLE_MATCH_MODE", "AUTHZ_SCOPE_MATCH_MODE",
		"AUTHZ_ROLE_CLAIM_PATHS", "AUTHZ_SCOPE_CLAIM_PATHS",
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
	if cfg.AuthIssuer != "" || cfg.AuthAudience != "" || cfg.AuthJWKSURL != "" || cfg.AuthIntrospectionURL != "" || cfg.AuthIntrospectionClientID != "" || cfg.AuthIntrospectionClientSecret != "" || cfg.AuthIntrospectionPrivateKey != "" || cfg.AuthIntrospectionPrivateKeyFile != "" || cfg.AuthIntrospectionPrivateKeyJWTKeyID != "" || cfg.AuthIntrospectionPrivateKeyJWTAlgorithm != "" {
		t.Errorf("expected auth config defaults to be empty, got %+v", cfg)
	}
	if len(cfg.AuthzRequiredRoles) != 0 || len(cfg.AuthzRequiredScopes) != 0 || len(cfg.AuthzRoleClaimPaths) != 0 || len(cfg.AuthzScopeClaimPaths) != 0 {
		t.Errorf("expected authz list defaults to be empty, got %+v", cfg)
	}
	if cfg.AuthzRoleMatchMode != "any" {
		t.Errorf("expected AuthzRoleMatchMode to default to 'any', got %s", cfg.AuthzRoleMatchMode)
	}
	if cfg.AuthzScopeMatchMode != "any" {
		t.Errorf("expected AuthzScopeMatchMode to default to 'any', got %s", cfg.AuthzScopeMatchMode)
	}
	if cfg.AuthTokenType != "jwt" {
		t.Errorf("expected AuthTokenType to default to 'jwt', got %s", cfg.AuthTokenType)
	}
	if cfg.AuthIntrospectionAuthMethod != "client_secret_basic" {
		t.Errorf("expected AuthIntrospectionAuthMethod to default to 'client_secret_basic', got %s", cfg.AuthIntrospectionAuthMethod)
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
