# S1.1: JWT Authentication Requirements Analysis

**Task**: Analyze JWT/token validation requirements
**Status**: 🔄 In Progress
**Date**: 2025-11-02
**Effort**: 2 hours (S1.1)
**Owner**: TBD

---

## Executive Summary

The C8S API server currently has **no token validation**. Any bearer token is accepted and converted to a hardcoded user with ID "user-id". This analysis defines the requirements for implementing proper JWT token validation to move toward production readiness.

---

## Current State Analysis

### Existing Code
**File**: `cmd/api-server/handlers/auth_middleware.go`

**Current Flow**:
```go
1. Extract token from Authorization header or auth_token cookie
2. If token exists (any token):
   - Create User object with hardcoded values:
     * ID: "user-id"
     * Username: "user"
     * Email: ""
     * Namespace: "default"
     * Roles: []string{}
3. Attach to context
4. No validation of token format, signature, or expiration
```

**Problems**:
- ✗ Any string is accepted as a token
- ✗ No signature verification
- ✗ No expiration checking
- ✗ No claims extraction
- ✗ Hardcoded user namespace
- ✗ No roles/permissions
- ✗ Security test at lines 54, 87: `// TODO: Integrate with C8S auth system`

### Code Locations
- **Main middleware**: `cmd/api-server/handlers/auth_middleware.go:54-64`
- **Optional auth variant**: `cmd/api-server/handlers/auth_middleware.go:87-94`
- **API Server startup**: `cmd/api-server/main.go` (no auth config)

### Dependencies Available
From `go.mod`:
- ✓ `golang.org/x/crypto` (crypto/rsa, etc. available)
- ✓ `golang.org/x/oauth2` (for OAuth2 support if needed)
- ✗ No JWT library currently in dependencies (needs to be added)

---

## Authentication Requirements

### 1. Token Format & Standards
**Decision**: Use **JWT (JSON Web Tokens)** for the following reasons:
- Industry standard for REST APIs
- Self-contained (no DB lookup for validation)
- Stateless authentication (good for Kubernetes)
- Wide library support across languages
- Can include metadata (claims)

**Supported Algorithms**:
- Primary: `HS256` (HMAC SHA-256) - for development
- Primary: `RS256` (RSA SHA-256) - for production with key rotation

**Token Structure**:
```
Header.Payload.Signature

Header: {
  "alg": "HS256" or "RS256",
  "typ": "JWT"
}

Payload (Claims): {
  "sub": "user-id",              # Subject (user ID)
  "name": "username",             # Username
  "email": "user@example.com",    # Email
  "namespace": "default",         # K8s namespace
  "roles": ["role1", "role2"],    # Roles/permissions
  "exp": 1234567890,              # Expiration time
  "iat": 1234567800,              # Issued at time
  "iss": "c8s-auth",              # Issuer
  "aud": "c8s-api"                # Audience
}

Signature: HMAC-SHA256(base64url(header) + "." + base64url(payload), secret)
```

---

### 2. Key Management

#### Option A: HS256 (Symmetric - Development Only)
```
Secret Key: Shared between auth provider and C8S API
Location: Environment variable JWT_SECRET
Size: 32-64 bytes (minimum for security)
```

**Pros**: Simple, fast, no key management
**Cons**: Less secure, requires secure key distribution, key rotation harder
**Use**: Development and single-organization deployments

**Implementation**:
```bash
# Generate secret (base64 encoded 32 bytes)
openssl rand -base64 32

# Set environment
export JWT_SECRET="base64-encoded-secret-here"
```

#### Option B: RS256 (Asymmetric - Production)
```
Public Key: Published by auth provider, used by C8S to verify
Private Key: Kept secret by auth provider, used to sign tokens
Format: PEM-encoded RSA keys (2048 or 4096 bits)
```

**Pros**: Industry standard, supports key rotation, auth provider keeps private key
**Cons**: Slightly slower, key management needed
**Use**: Production, multi-organization, external auth providers

**Implementation**:
```bash
# Public key provided by auth provider (typically via JWKS endpoint)
# https://auth-provider.com/.well-known/jwks.json

# C8S fetches and caches public key
# Or load from file: JWT_PUBLIC_KEY_PATH=/path/to/public.pem
```

---

### 3. Token Validation Checklist

When validating a token, check:

**✓ Signature Verification**
- Verify token signature using configured key/algorithm
- Reject invalid signatures
- Reject unknown algorithms

**✓ Standard Claims**
- `exp` (expiration): Token must not be expired
- `iat` (issued at): Must be in past
- `iss` (issuer): Must match expected issuer
- `aud` (audience): Must match expected audience (optional but recommended)

**✓ Custom Claims** (extracted to User struct)
- `sub`: User ID (required)
- `name`: Username (required)
- `email`: Email (optional)
- `namespace`: K8s namespace (required)
- `roles`: List of roles/permissions (optional)

**✓ Token Format**
- Must be Bearer token: `Authorization: Bearer <token>`
- Token must be properly formatted JWT
- Must have exactly 3 parts separated by dots

---

### 4. Configuration Options

#### Environment Variables
```bash
# Authentication Mode
AUTH_MODE=jwt                      # "jwt", "none" (dev only), "oauth2" (future)

# JWT Configuration
JWT_ALGORITHM=HS256                # "HS256" or "RS256"
JWT_ISSUER=c8s-auth               # Expected issuer
JWT_AUDIENCE=c8s-api              # Expected audience

# For HS256 (symmetric key)
JWT_SECRET=base64-encoded-secret   # 32-64 bytes, base64 encoded

# For RS256 (asymmetric key)
JWT_PUBLIC_KEY_PATH=/path/to/key   # Path to PEM public key
JWT_JWKS_URL=https://auth/.../jwks # JWKS endpoint (alternative to file)
JWT_PUBLIC_KEY_CACHE_TTL=3600      # Cache TTL in seconds

# Token Validation
JWT_EXPIRY_TOLERANCE=60            # Clock skew tolerance in seconds
JWT_VERIFY_EXPIRY=true             # Whether to check expiration
JWT_VERIFY_SIGNATURE=true          # Whether to verify signature

# Default User Claims (if missing from token)
JWT_DEFAULT_NAMESPACE=default      # Fallback namespace
JWT_DEFAULT_ROLES=user,viewer      # Comma-separated default roles
```

---

### 5. Error Handling Strategy

### Token Validation Errors
```go
// Error codes to return
401 Unauthorized - Invalid or missing token
401 Unauthorized - Token has expired
401 Unauthorized - Invalid token signature
401 Unauthorized - Token missing required claims
400 Bad Request - Malformed token
```

### Security Considerations
- **Don't leak token details**: Don't expose why token is invalid
- **Log for analysis**: Log validation failures to security logs
- **Rate limiting**: Consider rate limiting failed auth attempts
- **Token blacklist** (optional): For token revocation before expiry

---

## Implementation Plan

### Approach
1. **Phase 1**: Add JWT library dependency
2. **Phase 2**: Create JWT validation package
3. **Phase 3**: Update middleware to use JWT validation
4. **Phase 4**: Add configuration loading
5. **Phase 5**: Add unit tests
6. **Phase 6**: Document (S1.6)

### Libraries to Add
```go
// Two viable options:

// Option 1: golang-jwt/jwt (most popular, v5+)
github.com/golang-jwt/jwt/v5 v5.x

// Option 2: lestrrat-go/jwx (more features, used by big projects)
github.com/lestrrat-go/jwx/v2 v2.x

// Recommendation: golang-jwt/jwt v5 (simpler, well-maintained)
```

### New Files to Create
```
cmd/api-server/
├─ auth/
│  ├─ jwt_validator.go          # JWT validation logic
│  ├─ jwt_claims.go             # Claims struct definition
│  └─ config.go                 # Auth configuration
│
tests/unit/
└─ auth/
   ├─ jwt_validator_test.go     # JWT validation tests
   └─ config_test.go            # Config loading tests
```

### Modified Files
```
cmd/api-server/
├─ handlers/
│  └─ auth_middleware.go        # Use JWT validator (S1.2)
├─ main.go                      # Load auth config
└─ go.mod / go.sum             # Add JWT dependency
```

---

## Detailed Design Decisions

### 1. Claims Structure
```go
// Standard JWT claims + custom C8S claims
type Claims struct {
    // Standard JWT claims (RFC 7519)
    Subject   string   `json:"sub"`  // User ID
    Name      string   `json:"name"` // Username
    Email     string   `json:"email"`
    ExpiresAt int64    `json:"exp"`
    IssuedAt  int64    `json:"iat"`
    Issuer    string   `json:"iss"`
    Audience  []string `json:"aud"`

    // C8S-specific claims
    Namespace string   `json:"namespace"`
    Roles     []string `json:"roles"`

    jwt.RegisteredClaims // Embed for standard validation
}
```

### 2. User Extraction
```go
// Current: Hardcoded
user := &User{
    ID:        "user-id",
    Username:  "user",
    Namespace: "default",
    Roles:     []string{},
}

// After: From JWT claims
user := &User{
    ID:        claims.Subject,
    Username:  claims.Name,
    Email:     claims.Email,
    Namespace: claims.Namespace,
    Roles:     claims.Roles,
}
```

### 3. Namespace Validation
```go
// Should C8S validate the namespace?
// Option A: Accept namespace from token (trust auth provider)
// Option B: Look up user -> namespace mapping in K8s
// Option C: Use hardcoded namespace (current approach)

// Recommendation: Option A initially, Option B in Phase 2
// Auth provider is responsible for correct namespace assignment
```

### 4. Token Sources (Priority Order)
```go
1. Authorization header: "Authorization: Bearer <token>"
2. Cookie: "auth_token" (for development/UI)
3. Query parameter: "?token=<token>" (NOT RECOMMENDED - logged in URLs)

// Implementation: Check in order above, use first found
```

---

## Testing Strategy

### Unit Tests to Write (S1.5)
1. **Valid Token Tests**
   - HS256 token with valid signature
   - RS256 token with valid signature
   - Token with all claims
   - Token with minimal claims

2. **Invalid Token Tests**
   - Missing signature
   - Wrong signature
   - Expired token
   - Token from different issuer
   - Wrong audience
   - Malformed token (not valid JWT)
   - Empty token

3. **Claims Extraction Tests**
   - Extract user from claims
   - Handle missing optional claims
   - Default values applied correctly
   - Roles parsed correctly

4. **Error Cases**
   - No token provided
   - Malformed Authorization header
   - Unknown algorithm
   - Key loading failure

### Integration Tests (Phase 2)
1. Full authentication flow with valid token
2. Middleware rejects invalid tokens
3. User object created correctly
4. Context contains user

---

## Configuration Recommendations

### Development Environment
```bash
# Simple HS256 development setup
AUTH_MODE=jwt
JWT_ALGORITHM=HS256
JWT_ISSUER=c8s-auth-dev
JWT_AUDIENCE=c8s-api
JWT_SECRET=$(openssl rand -base64 32)
JWT_VERIFY_EXPIRY=false            # Allow expired tokens in dev
```

### Production Environment
```bash
# Secure RS256 production setup
AUTH_MODE=jwt
JWT_ALGORITHM=RS256
JWT_ISSUER=https://your-auth-provider.com
JWT_AUDIENCE=c8s-api
JWT_JWKS_URL=https://your-auth-provider.com/.well-known/jwks.json
JWT_VERIFY_EXPIRY=true
JWT_VERIFY_SIGNATURE=true
```

### Token Example (Development)
```json
{
  "alg": "HS256",
  "typ": "JWT"
}.{
  "sub": "user-123",
  "name": "john-doe",
  "email": "john@example.com",
  "namespace": "default",
  "roles": ["admin", "viewer"],
  "iss": "c8s-auth-dev",
  "aud": ["c8s-api"],
  "iat": 1699041600,
  "exp": 1699128000
}
```

---

## Security Considerations

### 1. Token Storage
- **Never store secrets in code** ✓ Use environment variables
- **Use secure channels** ✓ HTTPS/TLS required in production
- **Don't log tokens** ✓ Sanitize logs
- **Short expiration** ✓ Recommend 1-24 hours

### 2. Key Management
- **HS256 Secret**: Min 32 bytes, rotate periodically
- **RS256 Private Key**: Never exposed, rotate keys regularly
- **Public Key**: Can be cached, validate against JWKS
- **Key rotation**: Plan for algorithm/key transitions

### 3. Attack Prevention
- **Signature bypass**: Always verify signature
- **Expiration bypass**: Always check expiration
- **Token reuse**: Validate aud (audience) claim
- **Token injection**: Use Bearer scheme validation
- **Timing attacks**: Use constant-time comparison

---

## Dependencies to Add

### Primary Choice: golang-jwt/jwt v5
```bash
go get github.com/golang-jwt/jwt/v5
```

**Why**:
- Most popular JWT library in Go
- Active maintenance and security updates
- Simple API: `jwt.ParseWithClaims(tokenString, &claims, keyFunc)`
- Good error handling
- Standard library compatibility

### Alternative: lestrrat-go/jwx/v2
```bash
go get github.com/lestrrat-go/jwx/v2
```

**When to use**:
- Need advanced features (JWE, JWS, etc.)
- Need JWKS support built-in
- Large-scale production system

**Recommendation**: Start with `golang-jwt/jwt/v5`, upgrade to `jwx` if needed

---

## Success Criteria for S1.1

✓ **Analysis Complete**:
- [x] Current authentication mechanism documented
- [x] Requirements clearly defined
- [x] JWT format and claims specified
- [x] Key management strategies outlined
- [x] Implementation approach designed
- [x] Testing strategy planned
- [x] Configuration documented
- [x] Security considerations addressed

✓ **Ready for S1.2**:
- [x] Team understands JWT requirements
- [x] Library choice decided (golang-jwt/jwt v5)
- [x] Error handling approach clear
- [x] Token format documented with examples
- [x] Integration points identified

---

## Next Steps (S1.2)

### Immediate Actions
1. ✅ Get team approval for JWT approach
2. ✅ Add `github.com/golang-jwt/jwt/v5` to go.mod
3. ✅ Create `cmd/api-server/auth/` package
4. ✅ Implement JWT validator
5. ✅ Write unit tests

### Code Outline for S1.2
```go
// cmd/api-server/auth/jwt_validator.go

package auth

import (
    "fmt"
    "time"
    jwt "github.com/golang-jwt/jwt/v5"
)

type JWTValidator struct {
    algorithm string
    issuer    string
    audience  string
    secret    string // For HS256
    publicKey interface{} // For RS256
}

type Claims struct {
    // Standard
    Subject string `json:"sub"`
    Name    string `json:"name"`
    Email   string `json:"email"`

    // C8S-specific
    Namespace string   `json:"namespace"`
    Roles     []string `json:"roles"`

    jwt.RegisteredClaims
}

func (v *JWTValidator) ValidateToken(tokenString string) (*Claims, error) {
    // 1. Parse token
    // 2. Validate signature
    // 3. Check claims (exp, iss, aud)
    // 4. Extract custom claims
    // 5. Return error or claims

    // TODO: Implement in S1.2
}
```

---

## Related Documents

- **Main Issue**: SYSTEMATIC_REVIEW.md - Section 2.1 "Authentication Not Implemented"
- **Implementation**: IMPROVEMENT_PLAN.md - Tasks S1.1 through S1.6
- **Progress**: PROGRESS_TRACKER.md - Track S1.x tasks
- **Quick Ref**: QUICK_REFERENCE.md - Command shortcuts

---

## Questions & Decisions Needed

| Question | Decision | Owner |
|----------|----------|-------|
| JWT Library: golang-jwt or lestrrat-go? | golang-jwt/jwt v5 | Tech Lead |
| Default algorithm: HS256 or RS256? | HS256 for dev, RS256 for prod | Architect |
| Token source: Header, cookie, query? | Header + Cookie | Lead Dev |
| Namespace validation approach? | Option A (trust provider) | Architect |
| Token expiration time? | 24 hours | Security |
| Key rotation strategy? | Document in S1.6 | Security |

---

**Analysis Status**: ✅ Complete
**Estimated Time Used**: 2 hours
**Next Task**: S1.2 - Implement JWT parsing and validation
**Ready to Proceed**: Yes

