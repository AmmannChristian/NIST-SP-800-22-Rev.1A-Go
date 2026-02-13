// Package config provides environment-variable-based configuration for the
// NIST SP 800-22 service, including gRPC port, TLS settings, Prometheus metrics
// port, log level, and OIDC authentication parameters.
package config

import (
	"crypto/tls"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultGRPCMaxMessageSize = 10 * 1024 * 1024

// Config holds all service configuration loaded from environment variables.
// Fields map one-to-one to environment variables (e.g. GRPCPort from GRPC_PORT).
type Config struct {
	// gRPC server configuration
	GRPCPort               int
	GRPCMaxRecvMessageSize int
	GRPCMaxSendMessageSize int

	// TLS configuration for gRPC
	TLSEnabled    bool
	TLSCertFile   string
	TLSKeyFile    string
	TLSCAFile     string
	TLSClientAuth string
	TLSMinVersion string

	// Metrics server configuration
	MetricsPort int

	// Logging configuration
	LogLevel string

	// Authentication configuration
	AuthEnabled                             bool
	AuthIssuer                              string
	AuthAudience                            string
	AuthJWKSURL                             string
	AuthTokenType                           string
	AuthIntrospectionURL                    string
	AuthIntrospectionAuthMethod             string
	AuthIntrospectionClientID               string
	AuthIntrospectionClientSecret           string
	AuthIntrospectionPrivateKey             string
	AuthIntrospectionPrivateKeyFile         string
	AuthIntrospectionPrivateKeyJWTKeyID     string
	AuthIntrospectionPrivateKeyJWTAlgorithm string
	AuthzRequiredRoles                      []string
	AuthzRequiredScopes                     []string
	AuthzRoleMatchMode                      string
	AuthzScopeMatchMode                     string
	AuthzRoleClaimPaths                     []string
	AuthzScopeClaimPaths                    []string
}

// Load reads configuration from environment variables with sensible defaults
// and validates the result. It returns an error if any required values are
// missing or out of range.
func Load() (*Config, error) {
	cfg := &Config{
		GRPCPort:                                getEnvInt("GRPC_PORT", 9090),
		GRPCMaxRecvMessageSize:                  getEnvInt("GRPC_MAX_RECV_MESSAGE_SIZE", defaultGRPCMaxMessageSize),
		GRPCMaxSendMessageSize:                  getEnvInt("GRPC_MAX_SEND_MESSAGE_SIZE", defaultGRPCMaxMessageSize),
		TLSEnabled:                              getEnvBool("TLS_ENABLED", false),
		TLSCertFile:                             getEnvString("TLS_CERT_FILE", ""),
		TLSKeyFile:                              getEnvString("TLS_KEY_FILE", ""),
		TLSCAFile:                               getEnvString("TLS_CA_FILE", ""),
		TLSClientAuth:                           getEnvString("TLS_CLIENT_AUTH", "none"),
		TLSMinVersion:                           getEnvString("TLS_MIN_VERSION", "1.2"),
		MetricsPort:                             getEnvInt("METRICS_PORT", 9091),
		LogLevel:                                getEnvString("LOG_LEVEL", "info"),
		AuthEnabled:                             getEnvBool("AUTH_ENABLED", false),
		AuthIssuer:                              getEnvString("AUTH_ISSUER", ""),
		AuthAudience:                            getEnvString("AUTH_AUDIENCE", ""),
		AuthJWKSURL:                             getEnvString("AUTH_JWKS_URL", ""),
		AuthTokenType:                           getEnvString("AUTH_TOKEN_TYPE", "jwt"),
		AuthIntrospectionURL:                    getEnvString("AUTH_INTROSPECTION_URL", ""),
		AuthIntrospectionAuthMethod:             getEnvString("AUTH_INTROSPECTION_AUTH_METHOD", "client_secret_basic"),
		AuthIntrospectionClientID:               getEnvString("AUTH_INTROSPECTION_CLIENT_ID", ""),
		AuthIntrospectionClientSecret:           getEnvString("AUTH_INTROSPECTION_CLIENT_SECRET", ""),
		AuthIntrospectionPrivateKey:             getEnvString("AUTH_INTROSPECTION_PRIVATE_KEY", ""),
		AuthIntrospectionPrivateKeyFile:         getEnvString("AUTH_INTROSPECTION_PRIVATE_KEY_FILE", ""),
		AuthIntrospectionPrivateKeyJWTKeyID:     getEnvString("AUTH_INTROSPECTION_PRIVATE_KEY_JWT_KID", ""),
		AuthIntrospectionPrivateKeyJWTAlgorithm: getEnvString("AUTH_INTROSPECTION_PRIVATE_KEY_JWT_ALG", ""),
		AuthzRequiredRoles:                      parseCSV(getEnvString("AUTHZ_REQUIRED_ROLES", "")),
		AuthzRequiredScopes:                     parseCSV(getEnvString("AUTHZ_REQUIRED_SCOPES", "")),
		AuthzRoleMatchMode:                      getEnvString("AUTHZ_ROLE_MATCH_MODE", "any"),
		AuthzScopeMatchMode:                     getEnvString("AUTHZ_SCOPE_MATCH_MODE", "any"),
		AuthzRoleClaimPaths:                     parseCSV(getEnvString("AUTHZ_ROLE_CLAIM_PATHS", "")),
		AuthzScopeClaimPaths:                    parseCSV(getEnvString("AUTHZ_SCOPE_CLAIM_PATHS", "")),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate checks that all configuration values are within acceptable ranges.
// It enforces that authentication credentials are present when AUTH_ENABLED is
// true and that TLS certificate paths are set when TLS_ENABLED is true.
func (c *Config) Validate() error {
	if c.GRPCPort < 1 || c.GRPCPort > 65535 {
		return fmt.Errorf("invalid GRPC_PORT: %d (must be 1-65535)", c.GRPCPort)
	}
	if c.GRPCMaxRecvMessageSize < 0 {
		return fmt.Errorf("invalid GRPC_MAX_RECV_MESSAGE_SIZE: %d (must be >= 0)", c.GRPCMaxRecvMessageSize)
	}
	if c.GRPCMaxSendMessageSize < 0 {
		return fmt.Errorf("invalid GRPC_MAX_SEND_MESSAGE_SIZE: %d (must be >= 0)", c.GRPCMaxSendMessageSize)
	}
	if c.GRPCMaxRecvMessageSize == 0 {
		c.GRPCMaxRecvMessageSize = defaultGRPCMaxMessageSize
	}
	if c.GRPCMaxSendMessageSize == 0 {
		c.GRPCMaxSendMessageSize = defaultGRPCMaxMessageSize
	}

	if c.MetricsPort < 1 || c.MetricsPort > 65535 {
		return fmt.Errorf("invalid METRICS_PORT: %d (must be 1-65535)", c.MetricsPort)
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[c.LogLevel] {
		return fmt.Errorf("invalid LOG_LEVEL: %s (must be debug/info/warn/error)", c.LogLevel)
	}

	roleMatchMode, err := parseAuthzMatchMode(c.AuthzRoleMatchMode, "AUTHZ_ROLE_MATCH_MODE")
	if err != nil {
		return err
	}
	c.AuthzRoleMatchMode = roleMatchMode

	scopeMatchMode, err := parseAuthzMatchMode(c.AuthzScopeMatchMode, "AUTHZ_SCOPE_MATCH_MODE")
	if err != nil {
		return err
	}
	c.AuthzScopeMatchMode = scopeMatchMode

	c.AuthzRequiredRoles = normalizeCSVValues(c.AuthzRequiredRoles)
	c.AuthzRequiredScopes = normalizeCSVValues(c.AuthzRequiredScopes)
	c.AuthzRoleClaimPaths = normalizeCSVValues(c.AuthzRoleClaimPaths)
	c.AuthzScopeClaimPaths = normalizeCSVValues(c.AuthzScopeClaimPaths)

	if c.AuthEnabled {
		if c.AuthIssuer == "" {
			return fmt.Errorf("invalid AUTH_ISSUER: required when AUTH_ENABLED=true")
		}
		if c.AuthAudience == "" {
			return fmt.Errorf("invalid AUTH_AUDIENCE: required when AUTH_ENABLED=true")
		}
		tokenType, err := parseAuthTokenType(c.AuthTokenType)
		if err != nil {
			return err
		}
		c.AuthTokenType = tokenType

		if c.AuthTokenType == "opaque" {
			if c.AuthIntrospectionURL == "" {
				return fmt.Errorf("invalid AUTH_INTROSPECTION_URL: required when AUTH_TOKEN_TYPE=opaque")
			}
			authMethod, parseErr := parseAuthIntrospectionAuthMethod(c.AuthIntrospectionAuthMethod)
			if parseErr != nil {
				return parseErr
			}
			c.AuthIntrospectionAuthMethod = authMethod

			switch c.AuthIntrospectionAuthMethod {
			case "client_secret_basic":
				if c.AuthIntrospectionClientID == "" {
					return fmt.Errorf("invalid AUTH_INTROSPECTION_CLIENT_ID: required when AUTH_INTROSPECTION_AUTH_METHOD=client_secret_basic")
				}
				if c.AuthIntrospectionClientSecret == "" {
					return fmt.Errorf("invalid AUTH_INTROSPECTION_CLIENT_SECRET: required when AUTH_INTROSPECTION_AUTH_METHOD=client_secret_basic")
				}
			case "private_key_jwt":
				c.AuthIntrospectionPrivateKey = strings.TrimSpace(c.AuthIntrospectionPrivateKey)
				c.AuthIntrospectionPrivateKeyFile = strings.TrimSpace(c.AuthIntrospectionPrivateKeyFile)
				if c.AuthIntrospectionPrivateKey != "" && c.AuthIntrospectionPrivateKeyFile != "" {
					return fmt.Errorf("invalid AUTH_INTROSPECTION_PRIVATE_KEY configuration: AUTH_INTROSPECTION_PRIVATE_KEY and AUTH_INTROSPECTION_PRIVATE_KEY_FILE are mutually exclusive")
				}
				if c.AuthIntrospectionPrivateKey == "" && c.AuthIntrospectionPrivateKeyFile == "" {
					return fmt.Errorf("invalid AUTH_INTROSPECTION_PRIVATE_KEY: required when AUTH_INTROSPECTION_AUTH_METHOD=private_key_jwt")
				}
				if c.AuthIntrospectionPrivateKeyFile != "" {
					privateKeyBytes, readErr := os.ReadFile(c.AuthIntrospectionPrivateKeyFile)
					if readErr != nil {
						return fmt.Errorf("invalid AUTH_INTROSPECTION_PRIVATE_KEY_FILE: %w", readErr)
					}
					c.AuthIntrospectionPrivateKey = strings.TrimSpace(string(privateKeyBytes))
					if c.AuthIntrospectionPrivateKey == "" {
						return fmt.Errorf("invalid AUTH_INTROSPECTION_PRIVATE_KEY_FILE: empty file")
					}
				}

				privateKeyJWTAlgorithm, algErr := parseAuthIntrospectionPrivateKeyJWTAlgorithm(c.AuthIntrospectionPrivateKeyJWTAlgorithm)
				if algErr != nil {
					return algErr
				}
				c.AuthIntrospectionPrivateKeyJWTAlgorithm = privateKeyJWTAlgorithm
			}
		}
	}

	if c.TLSEnabled {
		if c.TLSCertFile == "" {
			return fmt.Errorf("invalid TLS_CERT_FILE: required when TLS_ENABLED=true")
		}
		if c.TLSKeyFile == "" {
			return fmt.Errorf("invalid TLS_KEY_FILE: required when TLS_ENABLED=true")
		}
		if _, err := parseTLSClientAuth(c.TLSClientAuth); err != nil {
			return err
		}
		if _, err := parseTLSMinVersion(c.TLSMinVersion); err != nil {
			return err
		}
	}

	return nil
}

// TLSClientAuthType returns the parsed tls.ClientAuthType from configuration.
func (c *Config) TLSClientAuthType() (tls.ClientAuthType, error) {
	return parseTLSClientAuth(c.TLSClientAuth)
}

// TLSMinVersionValue returns the configured minimum TLS version (defaults to TLS 1.2).
func (c *Config) TLSMinVersionValue() (uint16, error) {
	return parseTLSMinVersion(c.TLSMinVersion)
}

// parseTLSClientAuth maps a human-readable client authentication mode string
// to the corresponding tls.ClientAuthType constant. The mapping is
// case-insensitive and accepts several aliases (e.g. "mtls" for
// RequireAndVerifyClientCert).
func parseTLSClientAuth(mode string) (tls.ClientAuthType, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "none", "noclientcert":
		return tls.NoClientCert, nil
	case "request", "requestclientcert":
		return tls.RequestClientCert, nil
	case "requireany", "requireanyclientcert":
		return tls.RequireAnyClientCert, nil
	case "verifyifgiven", "verify_client_cert_if_given":
		return tls.VerifyClientCertIfGiven, nil
	case "requireandverify", "requireandverifyclientcert", "mtls":
		return tls.RequireAndVerifyClientCert, nil
	default:
		return tls.NoClientCert, fmt.Errorf("invalid TLS_CLIENT_AUTH: %s", mode)
	}
}

// parseTLSMinVersion maps a version string (e.g. "1.2", "tls1.3") to the
// corresponding crypto/tls version constant. Only TLS 1.2 and 1.3 are supported.
func parseTLSMinVersion(version string) (uint16, error) {
	switch strings.ToLower(strings.TrimSpace(version)) {
	case "", "default", "1.2", "tls1.2", "tls12":
		return tls.VersionTLS12, nil
	case "1.3", "tls1.3", "tls13":
		return tls.VersionTLS13, nil
	default:
		return 0, fmt.Errorf("invalid TLS_MIN_VERSION: %s (use 1.2 or 1.3)", version)
	}
}

func parseAuthTokenType(tokenType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(tokenType)) {
	case "", "jwt":
		return "jwt", nil
	case "opaque":
		return "opaque", nil
	default:
		return "", fmt.Errorf("invalid AUTH_TOKEN_TYPE: %s (use jwt or opaque)", tokenType)
	}
}

func parseAuthIntrospectionAuthMethod(method string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case "", "client_secret_basic":
		return "client_secret_basic", nil
	case "private_key_jwt":
		return "private_key_jwt", nil
	default:
		return "", fmt.Errorf("invalid AUTH_INTROSPECTION_AUTH_METHOD: %s (use client_secret_basic or private_key_jwt)", method)
	}
}

func parseAuthIntrospectionPrivateKeyJWTAlgorithm(algorithm string) (string, error) {
	switch strings.ToUpper(strings.TrimSpace(algorithm)) {
	case "", "RS256", "ES256":
		return strings.ToUpper(strings.TrimSpace(algorithm)), nil
	default:
		return "", fmt.Errorf("invalid AUTH_INTROSPECTION_PRIVATE_KEY_JWT_ALG: %s (use RS256 or ES256)", algorithm)
	}
}

func parseAuthzMatchMode(mode string, envName string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "any":
		return "any", nil
	case "all":
		return "all", nil
	default:
		return "", fmt.Errorf("invalid %s: %s (use any or all)", envName, mode)
	}
}

func parseCSV(value string) []string {
	return normalizeCSVValues(strings.Split(value, ","))
}

func normalizeCSVValues(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	normalizedValues := make([]string, 0, len(values))
	for _, value := range values {
		normalizedValue := strings.TrimSpace(value)
		if normalizedValue == "" {
			continue
		}
		normalizedValues = append(normalizedValues, normalizedValue)
	}

	if len(normalizedValues) == 0 {
		return nil
	}

	return normalizedValues
}

// getEnvString reads a string from environment variable or returns default
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt reads an integer from environment variable or returns default
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool reads a boolean from environment variable or returns default
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
