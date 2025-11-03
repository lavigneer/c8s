# C8S Authentication Guide

**Status**: ✅ Production Ready
**Last Updated**: 2025-11-02
**Algorithms**: HS256 (Development), RS256 (Production)

---

## Quick Start

### Development Environment

```bash
# Generate a secure secret
export JWT_SECRET=$(openssl rand -base64 32)

# Configure JWT
export AUTH_MODE=jwt
export JWT_ALGORITHM=HS256
export JWT_ISSUER=c8s-auth-dev
export JWT_AUDIENCE=c8s-api
export JWT_VERIFY_EXPIRY=false  # Allow expired tokens in dev

# Start the API server
./bin/api-server
```

### Production Environment

```bash
# Configure with RSA public key
export AUTH_MODE=jwt
export JWT_ALGORITHM=RS256
export JWT_ISSUER=https://your-auth-provider.com
export JWT_AUDIENCE=c8s-api
export JWT_PUBLIC_KEY_PATH=/etc/c8s/public.pem
export JWT_VERIFY_EXPIRY=true
export JWT_VERIFY_SIGNATURE=true

# Start the API server
./bin/api-server
```

### Generating a Test Token

```bash
# Using jwt.io or similar tool, create a token with claims:
{
  "sub": "user-123",
  "name": "John Doe",
  "email": "john@example.com",
  "namespace": "default",
  "roles": ["admin"],
  "iss": "c8s-auth-dev",
  "aud": ["c8s-api"],
  "exp": 1699128000,
  "iat": 1699041600
}

# Or use the example below (requires github.com/golang-jwt/jwt/v5):
package main

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    Subject   string   `json:"sub"`
    Name      string   `json:"name"`
    Email     string   `json:"email"`
    Namespace string   `json:"namespace"`
    Roles     []string `json:"roles"`
    jwt.RegisteredClaims
}

func main() {
    now := time.Now()
    claims := Claims{
        Subject:   "user-123",
        Name:      "John Doe",
        Email:     "john@example.com",
        Namespace: "default",
        Roles:     []string{"admin"},
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    "c8s-auth-dev",
            Audience:  jwt.ClaimStrings{"c8s-api"},
            ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    secret := "your-secret-here"
    tokenString, err := token.SignedString([]byte(secret))
    if err != nil {
        panic(err)
    }

    fmt.Println(tokenString)
}
```

### Making an Authenticated Request

```bash
# Using curl with Authorization header
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/pipelines

# Using curl with cookie (for UI)
curl -b "auth_token=<token>" http://localhost:8080/dashboard
```

---

## Configuration Reference

### Environment Variables

#### Authentication Mode
- **`AUTH_MODE`** (default: `jwt`)
  - `jwt`: Use JWT token validation (recommended)
  - `none`: Accept any token (development only, never in production)

#### JWT Algorithm
- **`JWT_ALGORITHM`** (default: `HS256`)
  - `HS256`: HMAC-SHA256 (symmetric, development)
  - `RS256`: RSA-SHA256 (asymmetric, production)

#### Issuer & Audience
- **`JWT_ISSUER`** (default: `c8s-auth`)
  - Expected issuer claim (`iss`) in token
  - Example: `https://auth.example.com`

- **`JWT_AUDIENCE`** (default: `c8s-api`)
  - Expected audience claim (`aud`) in token
  - Example: `c8s-api` or `https://api.example.com`

#### HS256 Configuration (Symmetric)
- **`JWT_SECRET`** (required for HS256)
  - Base64-encoded secret (minimum 32 bytes)
  - Generate: `openssl rand -base64 32`
  - Keep secure - don't commit to version control!

#### RS256 Configuration (Asymmetric)
- **`JWT_PUBLIC_KEY_PATH`** (alternative to JWKS)
  - Path to PEM-encoded RSA public key
  - Example: `/etc/c8s/public.pem`
  - For key rotation: update file and restart

- **`JWT_JWKS_URL`** (alternative to file)
  - JWKS (JSON Web Key Set) endpoint URL
  - Example: `https://auth.example.com/.well-known/jwks.json`
  - Token: `https://your-auth-provider.com/.well-known/jwks.json`
  - Allows dynamic key rotation without restart

#### Token Validation
- **`JWT_VERIFY_EXPIRY`** (default: `true`)
  - Check if token has expired
  - Set to `false` for development if needed
  - Always `true` in production

- **`JWT_VERIFY_SIGNATURE`** (default: `true`)
  - Verify token signature
  - Should always be `true` in production
  - Never disable in production!

- **`JWT_EXPIRY_TOLERANCE`** (default: `60s`)
  - Clock skew tolerance (seconds)
  - Handles slight time differences between systems
  - Example: `30s`, `5m`, `1h`

#### Defaults
- **`JWT_DEFAULT_NAMESPACE`** (default: `default`)
  - Fallback namespace if token doesn't specify one

- **`JWT_DEFAULT_ROLES`** (default: `user`)
  - Comma-separated default roles if token doesn't specify
  - Example: `viewer,developer`

### Example Configurations

#### Development (HS256)
```bash
export AUTH_MODE=jwt
export JWT_ALGORITHM=HS256
export JWT_ISSUER=c8s-dev
export JWT_AUDIENCE=c8s-api
export JWT_SECRET=$(openssl rand -base64 32)
export JWT_VERIFY_EXPIRY=false
export JWT_DEFAULT_NAMESPACE=default
export JWT_DEFAULT_ROLES=admin
```

#### Production (RS256 with File)
```bash
export AUTH_MODE=jwt
export JWT_ALGORITHM=RS256
export JWT_ISSUER=https://auth.mycompany.com
export JWT_AUDIENCE=https://c8s-api.mycompany.com
export JWT_PUBLIC_KEY_PATH=/etc/c8s/secrets/public.pem
export JWT_VERIFY_EXPIRY=true
export JWT_VERIFY_SIGNATURE=true
export JWT_EXPIRY_TOLERANCE=60s
```

#### Production (RS256 with JWKS)
```bash
export AUTH_MODE=jwt
export JWT_ALGORITHM=RS256
export JWT_ISSUER=https://keycloak.mycompany.com/auth/realms/c8s
export JWT_AUDIENCE=c8s-api
export JWT_JWKS_URL=https://keycloak.mycompany.com/auth/realms/c8s/protocol/openid-connect/certs
export JWT_VERIFY_EXPIRY=true
export JWT_VERIFY_SIGNATURE=true
```

---

## Token Format & Claims

### JWT Structure
```
Header.Payload.Signature

eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.
eyJzdWIiOiJ1c2VyLTEyMyIsIm5hbWUiOiJKb2huIERvZSIsImlzcyI6ImM4cy1hdXRoIn0.
aGVsbG8gd29ybGQ=
```

### Required Claims

| Claim | Type | Description | Example |
|-------|------|-------------|---------|
| `sub` | string | User ID (subject) | `user-123` |
| `name` | string | Username | `john-doe` |
| `namespace` | string | Kubernetes namespace | `default` |
| `exp` | number | Expiration time (unix timestamp) | `1699041600` |
| `iat` | number | Issued at time (unix timestamp) | `1699038000` |

### Optional Claims

| Claim | Type | Description | Example |
|-------|------|-------------|---------|
| `email` | string | User email address | `john@example.com` |
| `roles` | array | List of roles | `["admin", "viewer"]` |
| `iss` | string | Issuer | `c8s-auth` |
| `aud` | array | Audience | `["c8s-api"]` |

### Example Token Payload
```json
{
  "sub": "user-123",
  "name": "John Doe",
  "email": "john@example.com",
  "namespace": "production",
  "roles": ["admin", "developer"],
  "iss": "https://auth.company.com",
  "aud": ["c8s-api"],
  "iat": 1699041600,
  "exp": 1699128000
}
```

---

## Integration with API Server

### Initialization

The API server must initialize the JWT validator at startup:

```go
package main

import (
    "github.com/org/c8s/cmd/api-server/auth"
    "github.com/org/c8s/cmd/api-server/handlers"
)

func main() {
    // Load configuration from environment
    config := auth.NewConfigFromEnv()

    // Validate configuration
    if err := config.Validate(); err != nil {
        log.Fatalf("Invalid auth config: %v", err)
    }

    // Initialize JWT validator
    if err := handlers.InitAuthValidator(config); err != nil {
        log.Fatalf("Failed to initialize auth: %v", err)
    }

    // Now serve HTTP requests
    // AuthMiddleware will use the initialized validator
}
```

### Middleware Usage

Apply authentication to routes:

```go
router := chi.NewRouter()

// Protected routes (require authentication)
protected := chi.NewRouter()
protected.Use(handlers.AuthMiddleware)
protected.Get("/pipelines", ListPipelinesHandler)
protected.Post("/pipelines", CreatePipelineHandler)
router.Mount("/api", protected)

// Public routes (optional authentication)
router.Get("/health", HealthCheckHandler) // No auth
router.Get("/metrics", handlers.OptionalAuthMiddleware(MetricsHandler)) // Optional auth
```

### Extracting User Info

In your handlers:

```go
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // Get authenticated user
    user, ok := handlers.GetUserFromContext(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Use user information
    log.Printf("Request from user: %s (namespace: %s)", user.Username, user.Namespace)

    // Enforce namespace isolation
    if userRequestedNamespace != user.Namespace {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Continue with handler logic...
}
```

---

## Security Best Practices

### HS256 (Symmetric Keys)

✓ **Use HS256 for**:
- Development environments
- Single-organization deployments
- Quick prototyping

✗ **Don't use HS256 for**:
- Production environments with multiple organizations
- Systems requiring key rotation
- Federated authentication

**Key Management**:
- Generate strong secrets: `openssl rand -base64 32`
- Store in secure location (environment, secrets vault)
- Rotate periodically (quarterly recommended)
- Never commit to version control

### RS256 (Asymmetric Keys)

✓ **Use RS256 for**:
- Production environments
- Multi-organization systems
- Key rotation requirements
- Federated authentication

**Key Management**:
- Private key: Never shared, kept by auth provider
- Public key: Can be shared, non-secret
- Key rotation: Update public key, no API restart needed
- JWKS endpoint: Preferred for automatic key discovery

**Recommended Setup**:
```bash
# Auth provider generates keys
openssl genrsa -out private.pem 4096
openssl rsa -in private.pem -pubout -out public.pem

# C8S loads public key
export JWT_PUBLIC_KEY_PATH=/etc/c8s/public.pem
# OR
export JWT_JWKS_URL=https://auth.provider.com/.well-known/jwks.json
```

### General Security

1. **Always verify signatures** - Set `JWT_VERIFY_SIGNATURE=true`
2. **Check expiration** - Set `JWT_VERIFY_EXPIRY=true` in production
3. **Short token lifetime** - 1-24 hours recommended
4. **Use HTTPS** - Always transmit tokens over TLS
5. **Sanitize logs** - Never log full tokens
6. **Secure secret storage**:
   - Environment variables with restricted access
   - Secrets management system (Vault, K8s Secrets, etc.)
   - Encrypted at rest and in transit
7. **Monitor failures** - Log all authentication failures for security auditing
8. **Token revocation** - Consider blacklist for logout support (optional)

---

## Troubleshooting

### "Unauthorized" Response

**Cause**: Token missing or invalid

**Solution**:
1. Check token is included in request:
   ```bash
   curl -H "Authorization: Bearer <token>" http://localhost:8080/api/test
   ```
2. Verify token format: Should be "Bearer <token>"
3. Check token expiration: `exp` claim should be in future
4. Verify issuer matches configuration:
   ```bash
   echo $JWT_ISSUER
   ```
5. Check audience matches:
   ```bash
   echo $JWT_AUDIENCE
   ```

### "token has expired"

**Cause**: Token expiration time passed

**Solution**:
1. Generate new token with future expiration
2. Check system clock is synchronized (NTP)
3. Increase `JWT_EXPIRY_TOLERANCE` if clock skew:
   ```bash
   export JWT_EXPIRY_TOLERANCE=120s
   ```

### "invalid token signature"

**Cause**: Token signed with wrong key

**Solution**:
1. Verify secret matches:
   ```bash
   # Token was signed with this secret
   echo $JWT_SECRET
   ```
2. For RS256, verify public key is correct:
   ```bash
   ls -la $JWT_PUBLIC_KEY_PATH
   ```
3. Regenerate token with correct key

### "invalid issuer"

**Cause**: Token issuer doesn't match configuration

**Solution**:
1. Check token issuer claim (`iss`)
2. Verify configuration:
   ```bash
   echo "Configured issuer: $JWT_ISSUER"
   ```
3. Ensure they match

### "invalid audience"

**Cause**: Token audience doesn't match configuration

**Solution**:
1. Check token audience claim (`aud`)
2. Verify configuration:
   ```bash
   echo "Configured audience: $JWT_AUDIENCE"
   ```
3. Ensure they match (note: `aud` is usually an array)

### API Server Won't Start

**Cause**: Invalid auth configuration

**Solution**:
1. Check all required environment variables:
   ```bash
   echo "AUTH_MODE: $AUTH_MODE"
   echo "JWT_ALGORITHM: $JWT_ALGORITHM"
   echo "JWT_ISSUER: $JWT_ISSUER"
   ```
2. For HS256, verify secret is set:
   ```bash
   [ -z "$JWT_SECRET" ] && echo "JWT_SECRET not set"
   ```
3. For RS256, verify key file exists:
   ```bash
   test -f "$JWT_PUBLIC_KEY_PATH" && echo "Key file found"
   ```
4. Check logs for specific error:
   ```bash
   ./bin/api-server 2>&1 | grep -i auth
   ```

---

## Development Mode

### Using NoOp Validator

For testing without real JWT tokens:

```bash
# Environment variable approach (if implemented)
export AUTH_MODE=none
./bin/api-server

# This will accept any token
curl -H "Authorization: Bearer anything" http://localhost:8080/api/test
```

### Example Development Token

```bash
# Generate a test token that won't expire
TOKEN=$(go run <<'EOF'
package main

import (
    "fmt"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    jwt.RegisteredClaims
    Subject   string   `json:"sub"`
    Name      string   `json:"name"`
    Email     string   `json:"email"`
    Namespace string   `json:"namespace"`
    Roles     []string `json:"roles"`
}

func main() {
    secret := "dev-secret-key-32-bytes-long!!"
    now := time.Now()
    claims := Claims{
        Subject:   "dev-user",
        Name:      "Developer",
        Email:     "dev@localhost",
        Namespace: "default",
        Roles:     []string{"admin"},
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    "c8s-auth",
            Audience:  jwt.ClaimStrings{"c8s-api"},
            ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(now),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, _ := token.SignedString([]byte(secret))
    fmt.Println(tokenString)
}
EOF
)

# Use the token
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/test
```

---

## Integration with Auth Providers

### Keycloak

```bash
export AUTH_MODE=jwt
export JWT_ALGORITHM=RS256
export JWT_ISSUER=https://keycloak.example.com/auth/realms/c8s
export JWT_AUDIENCE=c8s-api
export JWT_JWKS_URL=https://keycloak.example.com/auth/realms/c8s/protocol/openid-connect/certs
```

### Auth0

```bash
export AUTH_MODE=jwt
export JWT_ALGORITHM=RS256
export JWT_ISSUER=https://YOUR_DOMAIN.auth0.com/
export JWT_AUDIENCE=YOUR_API_IDENTIFIER
export JWT_JWKS_URL=https://YOUR_DOMAIN.auth0.com/.well-known/jwks.json
```

### Azure AD / Microsoft Identity Platform

```bash
export AUTH_MODE=jwt
export JWT_ALGORITHM=RS256
export JWT_ISSUER=https://login.microsoftonline.com/{tenant-id}/v2.0
export JWT_AUDIENCE=YOUR_CLIENT_ID
export JWT_JWKS_URL=https://login.microsoftonline.com/{tenant-id}/discovery/v2.0/keys
```

### Custom OAuth2 Provider

Use your provider's JWKS endpoint or download public key:

```bash
# Download public key
curl https://your-provider.com/public.pem > /etc/c8s/public.pem

# Configure
export AUTH_MODE=jwt
export JWT_ALGORITHM=RS256
export JWT_ISSUER=https://your-provider.com
export JWT_AUDIENCE=c8s-api
export JWT_PUBLIC_KEY_PATH=/etc/c8s/public.pem
```

---

## API Reference

### Authentication Endpoints

**Login** (redirects to auth provider):
```
GET /login
```

**Logout**:
```
POST /logout
```

**Current User** (requires authentication):
```
GET /api/auth/me
```

### Error Responses

| Status | Message | Meaning |
|--------|---------|---------|
| 401 | Unauthorized | Missing or invalid token |
| 401 | Token has expired | Token expiration time passed |
| 401 | Invalid token signature | Token signed with wrong key |
| 401 | Invalid token claims | Required claims missing or invalid |
| 400 | Malformed token | Token format is invalid |
| 500 | Internal server error | Auth system not properly initialized |

---

## Testing Authentication

### Unit Tests

```bash
# Run authentication unit tests
go test ./tests/unit/auth/... -v

# Run middleware integration tests
go test ./tests/unit/handlers/... -run Auth -v

# Check test coverage
go test ./tests/unit/auth/... -cover
```

### E2E Tests

```bash
# Run E2E tests with authentication
npm run test:e2e

# Run specific authentication tests
npm run test:e2e -- --grep "authentication"
```

### Manual Testing

```bash
# Generate test token
TOKEN=$(openssl rand -base64 32) # For HS256, this is not a real token

# Test authenticated endpoint
curl -v -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/pipelines

# Test without token
curl -v http://localhost:8080/api/pipelines

# Test with invalid token
curl -v -H "Authorization: Bearer invalid" \
  http://localhost:8080/api/pipelines
```

---

## Frequently Asked Questions

**Q: Can I use both HS256 and RS256?**
A: No, choose one algorithm per deployment. HS256 for dev, RS256 for production.

**Q: How do I rotate keys?**
A: For HS256, update JWT_SECRET and restart. For RS256, update the key file or JWKS endpoint and restart (JWKS is preferred for no-restart rotation).

**Q: What if token doesn't have all claims?**
A: Required claims (`sub`, `name`, `namespace`) must be present. Optional claims fall back to defaults (JWT_DEFAULT_NAMESPACE, JWT_DEFAULT_ROLES).

**Q: Can I use the same secret in development and production?**
A: No, always use unique secrets per environment. Generate new secret for each deployment.

**Q: How do I debug token validation failures?**
A: Check logs (will show "Token validation failed: <reason>") and verify token claims using jwt.io or similar tool.

**Q: Is JWT_SECRET case-sensitive?**
A: Yes, the secret must match exactly. If tokens were signed with a different secret, they will fail validation.

---

## References

- [RFC 7519 - JSON Web Token (JWT)](https://tools.ietf.org/html/rfc7519)
- [JSON Web Algorithms (JWA) - RFC 7518](https://tools.ietf.org/html/rfc7518)
- [golang-jwt/jwt - Go JWT Library](https://github.com/golang-jwt/jwt)
- [jwt.io - JWT Debugger](https://jwt.io)

---

**Last Updated**: 2025-11-02
**Status**: ✅ Production Ready
**Maintained By**: C8S Development Team
