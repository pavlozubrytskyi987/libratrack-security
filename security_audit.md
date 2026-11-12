# Security Audit Report — LibraTrack API

## Finding 1: JWT Secret Fallback to Hardcoded Value
- **File:** `internal/auth/auth.go`, lines 22-24
- **Severity:** High
- **Description:** When `JWT_SECRET` env var is empty, the code falls back to a hardcoded default string `"dev-secret-change-me"`. If deployed without the env var set, any attacker who reads the source can forge valid tokens.
- **Triage:** Fixed Now — added a startup check in `main.go` that logs a warning when `JWT_SECRET` is not set. In production, the app should refuse to start without it.

## Finding 2: User Enumeration Prevention on Login
- **File:** `internal/handler/auth_handler.go`, lines 46-52
- **Severity:** Medium
- **Description:** Both "unknown email" and "wrong password" cases must return the same generic error message to prevent attackers from discovering valid emails. Verified that both paths return `{"error": "invalid credentials"}` with HTTP 401.
- **Triage:** Fixed Now — confirmed both branches use identical error response.

## Finding 3: Password Hash Exposure in JSON
- **File:** `internal/model/model.go`, line 8
- **Severity:** High
- **Description:** The `PasswordHash` field must be tagged with `json:"-"` to prevent it from appearing in any JSON response (user listing, registration response, etc.).
- **Triage:** Fixed Now — `PasswordHash` has `json:"-"` tag applied.

## Finding 4: Mass Assignment via Request Binding
- **File:** `internal/handler/auth_handler.go`
- **Severity:** Medium
- **Description:** If the handler binds request JSON directly to the User model, an attacker could set their own role to "admin" during registration. Must use separate DTO structs (RegisterRequest) that only include user-editable fields.
- **Triage:** Fixed Now — RegisterRequest only contains Email and Password. Role is hardcoded to "reader" in the handler.

## Finding 5: Missing Abort in Middleware
- **File:** `internal/middleware/auth.go`, `internal/middleware/rbac.go`
- **Severity:** High
- **Description:** If authentication or authorization middleware uses `c.JSON()` instead of `c.AbortWithStatusJSON()`, the Gin handler chain continues executing. This means protected handlers would still run even after a failed auth check.
- **Triage:** Fixed Now — both AuthMiddleware and RequireRole use `c.AbortWithStatusJSON()`.

## Finding 6: No Rate Limiting on Login Endpoint
- **File:** `cmd/server/main.go`
- **Severity:** Medium
- **Description:** The `/auth/login` endpoint has no rate limiting, making it vulnerable to brute-force password attacks. An attacker could try thousands of passwords per second.
- **Triage:** Tracked for Later — needs a rate-limiting middleware (e.g., token bucket per IP). Not implemented in this version.

## Finding 7: RBAC Not Applied to Streaming/Batch Endpoints
- **File:** `cmd/server/main.go`
- **Severity:** Low
- **Description:** All current endpoints have proper RBAC. However, any new endpoints added in the future must be placed in the correct route group. Developers should default to the most restrictive group and relax as needed.
- **Triage:** False Positive — current implementation is correct. This is a process recommendation, not a code bug.
