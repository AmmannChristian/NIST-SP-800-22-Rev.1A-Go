# API Reference: NIST SP 800-22 Rev 1a Statistical Test Service

## Table of Contents

1. [Overview](#1-overview)
2. [gRPC Service Definition](#2-grpc-service-definition)
3. [Request Messages](#3-request-messages)
4. [Response Messages](#4-response-messages)
5. [Test Identifiers and Parameters](#5-test-identifiers-and-parameters)
6. [Error Handling](#6-error-handling)
7. [Health Check Service](#7-health-check-service)
8. [HTTP Endpoints](#8-http-endpoints)
9. [Prometheus Metrics Reference](#9-prometheus-metrics-reference)
10. [Authentication](#10-authentication)

---

## 1. Overview

This document provides a complete specification of the external interfaces exposed by the NIST SP 800-22 Rev 1a Statistical Test Service. The service exposes two network interfaces: a gRPC server for statistical test execution and an HTTP server for metrics, health checks, and performance profiling.

The gRPC interface is defined in the `nist.sp800_22.v1` Protocol Buffers package. The authoritative proto file resides at `api/nist/v1/nist_sp800_22.proto` within the service repository. Generated Go stubs are located at `pkg/pb/`.

## 2. gRPC Service Definition

### 2.1 Service: Sp80022TestService

The service provides a single unary RPC method for executing the complete NIST SP 800-22 test suite.

| Method | Request Type | Response Type | Description |
|---|---|---|---|
| `RunTestSuite` | `Sp80022TestRequest` | `Sp80022TestResponse` | Executes all 15 NIST SP 800-22 Rev 1a statistical tests on the provided bitstream |

**Service endpoint**: Default port 9090 (configurable via `GRPC_PORT`).

**Fully qualified method name**: `/nist.sp800_22.v1.Sp80022TestService/RunTestSuite`

## 3. Request Messages

### 3.1 Sp80022TestRequest

The primary request message containing the bitstream to test and optional configuration overrides.

```protobuf
message Sp80022TestRequest {
  bytes bitstream = 1;
  optional Sp80022TestConfig config = 2;
}
```

| Field | Number | Type | Required | Description |
|---|---|---|---|---|
| `bitstream` | 1 | `bytes` | Yes | Raw bitstream in packed byte format (big-endian bit order). Must contain between 387,840 and 10,000,000 bits (48,480 to 1,250,000 bytes). |
| `config` | 2 | `Sp80022TestConfig` | No | Per-test parameter overrides. If omitted, all tests use NIST-recommended default values. |

**Bitstream encoding**: Bits are packed in big-endian order within each byte. Bit index 0 corresponds to the most significant bit of byte 0. The total number of bits is computed as `len(bitstream) * 8`.

**Validation constraints**:
- The bitstream must not be empty.
- The total bit count must be at least 387,840 (the minimum required by the Universal Statistical Test).
- The total bit count must not exceed 10,000,000.

### 3.2 Sp80022TestConfig

Optional configuration message allowing the caller to override default test parameters.

```protobuf
message Sp80022TestConfig {
  int32 block_frequency_block_length = 1;
  int32 non_overlapping_template_block_length = 2;
  int32 overlapping_template_block_length = 3;
  int32 approximate_entropy_block_length = 4;
  int32 serial_block_length = 5;
  int32 linear_complexity_sequence_length = 6;
}
```

| Field | Number | Default | Valid Range | Test Affected |
|---|---|---|---|---|
| `block_frequency_block_length` | 1 | 128 | Positive integer, n/M >= 1 | Block Frequency Test |
| `non_overlapping_template_block_length` | 2 | 9 | Currently only 9 is supported | Non-Overlapping Template Matching Test |
| `overlapping_template_block_length` | 3 | 9 | Positive integer | Overlapping Template Matching Test |
| `approximate_entropy_block_length` | 4 | 10 | Positive integer, m >= 1 | Approximate Entropy Test |
| `serial_block_length` | 5 | 16 | Integer >= 2 | Serial Test |
| `linear_complexity_sequence_length` | 6 | 500 | Positive integer, n/M >= 1 | Linear Complexity Test |

When any field is set to zero or the entire config message is omitted, the corresponding default value is used.

## 4. Response Messages

### 4.1 Sp80022TestResponse

The aggregate response containing individual test results and summary statistics.

```protobuf
message Sp80022TestResponse {
  string timestamp = 1;
  int32 sample_size_bits = 2;
  double overall_pass_rate = 3;
  double p_value_uniformity_chi2 = 4;
  repeated Sp80022TestResult results = 5;
  int64 execution_time_ms = 6;
  int32 tests_run = 7;
  int32 tests_skipped = 8;
  int32 tests_total = 9;
  bool nist_compliant = 10;
}
```

| Field | Number | Type | Description |
|---|---|---|---|
| `timestamp` | 1 | `string` | ISO 8601 formatted timestamp (RFC 3339) of when the tests completed |
| `sample_size_bits` | 2 | `int32` | Total number of bits in the tested sample |
| `overall_pass_rate` | 3 | `double` | Fraction of executed (non-skipped) tests that passed, range [0.0, 1.0] |
| `p_value_uniformity_chi2` | 4 | `double` | P-value from the chi-squared uniformity test on individual p-values. Set to -1.0 if fewer than 5 valid p-values are available. |
| `results` | 5 | `repeated Sp80022TestResult` | Ordered list of individual test results (always 15 entries) |
| `execution_time_ms` | 6 | `int64` | Total wall-clock execution time in milliseconds |
| `tests_run` | 7 | `int32` | Number of tests that produced valid results |
| `tests_skipped` | 8 | `int32` | Number of tests that were skipped (e.g., insufficient data) |
| `tests_total` | 9 | `int32` | Total number of tests in the suite (always 15) |
| `nist_compliant` | 10 | `bool` | True if and only if `tests_run` equals `tests_total` |

**Pass rate computation**: The `overall_pass_rate` is computed only over tests that were actually executed (`tests_run`). Skipped tests do not factor into the denominator. If no tests were executed, the pass rate is 0.0.

**P-value uniformity**: The chi-squared test bins the individual p-values into 10 equal-width intervals and computes the goodness-of-fit statistic with 9 degrees of freedom. A p-value near 1.0 indicates that the individual p-values are uniformly distributed, which is the expected behaviour for truly random data.

### 4.2 Sp80022TestResult

The result of a single statistical test.

```protobuf
message Sp80022TestResult {
  string name = 1;
  double p_value = 2;
  bool passed = 3;
  optional double proportion = 4;
  optional string warning = 5;
}
```

| Field | Number | Type | Description |
|---|---|---|---|
| `name` | 1 | `string` | Snake-case test identifier (see Section 5) |
| `p_value` | 2 | `double` | Computed p-value in the range [0.0, 1.0]. A negative value indicates the test was skipped. |
| `passed` | 3 | `bool` | True if `p_value >= 0.01` (the NIST significance level Alpha) |
| `proportion` | 4 | `optional double` | 1.0 if the test passed, 0.0 if it failed. Present only for executed tests. |
| `warning` | 5 | `optional string` | Diagnostic message when the test could not complete normally (e.g., insufficient bits for the specific test). Empty when no issues occurred. |

## 5. Test Identifiers and Parameters

The following table lists all fifteen test identifiers as they appear in the `name` field of `Sp80022TestResult`, along with the default parameters used during execution.

| Index | Test Identifier | NIST Section | Default Parameters |
|---|---|---|---|
| 0 | `frequency_monobit` | 2.1 | None |
| 1 | `block_frequency` | 2.2 | M = 128 |
| 2 | `cumulative_sums` | 2.13 | None (forward + reverse, minimum reported) |
| 3 | `runs` | 2.3 | None |
| 4 | `longest_run` | 2.4 | Auto-selected M based on n |
| 5 | `binary_matrix_rank` | 2.5 | 32 x 32 matrices |
| 6 | `discrete_fourier_transform` | 2.6 | None |
| 7 | `non_overlapping_template` | 2.7 | m = 9, 148 templates, minimum p-value |
| 8 | `overlapping_template` | 2.8 | m = 9 |
| 9 | `universal_statistical` | 2.9 | L auto-selected based on n |
| 10 | `approximate_entropy` | 2.12 | m = 10 |
| 11 | `random_excursions` | 2.14 | States {-4,...,-1, 1,...,4}, minimum p-value |
| 12 | `random_excursions_variant` | 2.15 | States {-9,...,-1, 1,...,9}, minimum p-value |
| 13 | `serial` | 2.11 | m = 16 |
| 14 | `linear_complexity` | 2.10 | M = 500 |

## 6. Error Handling

The service returns standard gRPC error codes under the following conditions:

| Condition | Behaviour |
|---|---|
| Empty bitstream | Returns an error with message "bitstream cannot be empty" |
| Bitstream below 387,840 bits | Returns an error with message indicating the actual and required bit counts |
| Bitstream above 10,000,000 bits | Returns an error with message indicating the actual and maximum bit counts |
| Test execution failure | Returns an error wrapping the underlying failure |
| Authentication failure (when enabled) | Returns `UNAUTHENTICATED` status |

Individual tests that cannot complete due to data-specific conditions (e.g., insufficient cycles for Random Excursions) are not reported as errors. Instead, they return a p-value of 0.0 with `passed = false` and a descriptive `warning` string in the corresponding `Sp80022TestResult`.

## 7. Health Check Service

The service implements the standard gRPC health checking protocol as defined in `grpc.health.v1.Health`.

| Method | Description |
|---|---|
| `Check` | Returns `SERVING` when the service is operational |
| `Watch` | Streams health status changes |

Both methods are exempted from OIDC authentication when auth is enabled.

The serving status is set to `SERVING` immediately upon successful startup and registration of the NIST test service.

## 8. HTTP Endpoints

The HTTP server listens on the configured metrics port (default 9091) and exposes the following endpoints.

| Path | Method | Description |
|---|---|---|
| `/metrics` | GET | Prometheus metrics in text exposition format |
| `/health` | GET | JSON health check response |
| `/debug/pprof/` | GET | Go pprof profiling index page |
| `/debug/pprof/profile` | GET | CPU profile (accepts `?seconds=N` parameter) |
| `/debug/pprof/heap` | GET | Heap memory profile |
| `/debug/pprof/goroutine` | GET | Goroutine stack dumps |
| `/debug/pprof/threadcreate` | GET | Thread creation profile |
| `/debug/pprof/block` | GET | Block contention profile |
| `/debug/pprof/mutex` | GET | Mutex contention profile |
| `/debug/pprof/trace` | GET | Execution trace (accepts `?seconds=N` parameter) |

### 8.1 Health Check Response

The `/health` endpoint returns a JSON object with the following structure:

```json
{
  "status": "healthy",
  "version": "2.0.0"
}
```

| Field | Type | Description |
|---|---|---|
| `status` | string | Always "healthy" when the endpoint is reachable |
| `version` | string | Semantic version of the service (currently "2.0.0") |

## 9. Prometheus Metrics Reference

All metrics are registered under the `nist_` namespace.

### 9.1 nist_tests_total

- **Type**: Counter
- **Labels**: `test` (test identifier), `status` (`pass` or `fail`)
- **Description**: Cumulative count of individual statistical test executions.

### 9.2 nist_test_duration_seconds

- **Type**: Histogram
- **Labels**: `test` (test identifier)
- **Buckets**: Default Prometheus buckets
- **Description**: Wall-clock duration of individual statistical tests in seconds.

### 9.3 nist_overall_duration_seconds

- **Type**: Histogram
- **Labels**: None
- **Buckets**: Default Prometheus buckets
- **Description**: Wall-clock duration of the complete fifteen-test suite in seconds.

### 9.4 nist_last_overall_pass_rate

- **Type**: Gauge
- **Labels**: None
- **Description**: Most recently computed overall pass rate as a value between 0.0 and 1.0.

### 9.5 nist_p_value

- **Type**: Gauge
- **Labels**: `test` (test identifier)
- **Description**: Most recently observed p-value for each statistical test.

### 9.6 nist_requests_total

- **Type**: Counter
- **Labels**: `method` (gRPC method name), `status` (`success` or `error`)
- **Description**: Cumulative count of gRPC requests received by the service.

## 10. Authentication

When `AUTH_ENABLED=true`, the service requires a valid OIDC bearer token in the gRPC request metadata.

### 10.1 Token Requirements

| Parameter | Description |
|---|---|
| Transport | Bearer token in the `authorization` gRPC metadata header |
| Format | JSON Web Token (JWT) |
| Signature | Verified against the JWKS endpoint |
| Issuer claim (`iss`) | Must match the configured `AUTH_ISSUER` |
| Audience claim (`aud`) | Must match the configured `AUTH_AUDIENCE` |

### 10.2 Exempt Methods

The following gRPC methods are exempt from authentication:

- `/grpc.health.v1.Health/Check`
- `/grpc.health.v1.Health/Watch`

### 10.3 Response Headers

Every gRPC response includes the `x-request-id` metadata header containing a UUID v4 identifier for the request, regardless of authentication status.