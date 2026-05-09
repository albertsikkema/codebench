# Authorization

## Principle

Authorization determines what an authenticated user is allowed to do. Enforce it server-side on every request, at every layer — from API endpoints to individual resources. Default to deny. Client-side checks are cosmetic; server-side checks are security.

## Why

- **Authentication is not authorization**: Knowing *who* the user is (authentication) does not mean they can do *anything* they request (authorization). Every authenticated endpoint still needs permission checks.
- **Client-side checks are trivially bypassed**: Hiding a button in the UI does not prevent a user from calling the API endpoint directly. Every permission check that matters must exist on the server.
- **IDOR is the most common authorization vulnerability**: Insecure Direct Object References occur when an API accepts a resource ID from the client and returns the resource without checking if the requesting user owns or has access to it. This is consistently in the OWASP Top 10.

## Core Rules

### 1. Default-Deny

Every request is unauthorized unless explicitly permitted. New roles start with zero permissions. New endpoints are protected by default.

```
BAD:  Allow all, then add restrictions for sensitive operations
GOOD: Deny all, then grant specific permissions per role
```

This means:
- Adding a new endpoint without an authorization check = security vulnerability
- Adding a new role without assigning permissions = the role can do nothing (safe default)
- Forgetting to check permissions = request is rejected, not allowed

### 2. Enforce Server-Side on Every Request

Every API endpoint and server action must verify that the authenticated user has permission for the requested operation. No exceptions.

```
BAD:  Frontend hides the "Delete" button for non-admins.
      The DELETE endpoint has no server-side check.
      → Any user can call DELETE /api/resource/123 directly.

GOOD: Frontend hides the button (better UX).
      Server returns 403 if the user lacks delete permission.
      → Security does not depend on the client.
```

### 3. Check Object-Level Authorization

It's not enough to check "can this user access resources of type X." You must check "can this user access *this specific* resource."

```
BAD:  User is authenticated → allow GET /api/orders/456
      (user A can read user B's orders by guessing IDs)

GOOD: User is authenticated AND order 456 belongs to user → allow
      User is authenticated AND order 456 belongs to someone else → 403
```

Implement this with:
- **Ownership checks**: `WHERE user_id = current_user.id`
- **Team/org membership**: `WHERE org_id IN (user's org memberships)`
- **Explicit grants**: A permissions table linking users to specific resources

**Test with**: "User A tries to access User B's resource" — this test should exist for every resource endpoint.

### 4. Centralize Authorization Logic

Scatter permission checks across individual handlers and they will be inconsistent, incomplete, and hard to audit.

**Approaches** (pick one):
- **Middleware/decorator**: Check permissions before the handler runs, based on route metadata
- **Authorization service**: A dedicated module that handlers call with (user, action, resource) and get allow/deny
- **Policy engine**: For complex rules, use a policy framework (OPA, Casbin, Cedar)

The key requirement: a single place where you can answer "who can do what."

### 5. Use RBAC as the Default Model

Role-Based Access Control assigns users to roles, and roles have permissions. This covers most applications.

```
Roles:       admin, editor, viewer
Permissions: create, read, update, delete

admin  → create, read, update, delete
editor → create, read, update
viewer → read
```

**When to add ABAC**: When authorization depends on attributes beyond role — time of day, IP address, resource state, relationship to the resource owner. Layer ABAC on top of RBAC; don't replace RBAC entirely.

### 6. Protect Administrative Actions

Admin capabilities require extra safeguards:

- **Explicit assignment**: Admin roles are never self-assignable
- **Audit trail**: Every admin action is logged (who, what, when, from where)
- **Re-authentication**: Destructive admin operations (delete user, change permissions) require the admin to re-enter their password or complete MFA
- **Scoped admin**: Prefer scoped roles (org admin, project admin) over global superuser
- **Self-demotion prevention**: Admins must not be able to downgrade their own role. This prevents accidental lockout where no admin remains to restore access

### 7. Service-to-Service Authorization

When services call each other, use scoped credentials — not admin keys.

```
BAD:  Service A calls Service B with the global admin API key
      → If Service A is compromised, attacker has full access to Service B

GOOD: Service A has a service account with only the permissions it needs
      → Compromise of A is limited to A's granted permissions
```

## Return Codes

| Situation | Status code | Meaning |
|-----------|-------------|---------|
| No credentials provided | 401 Unauthorized | "Who are you?" |
| Credentials invalid/expired | 401 Unauthorized | "I don't recognize you" |
| Authenticated but not permitted | 403 Forbidden | "I know who you are, but you can't do this" |
| Resource not found OR not accessible | 404 Not Found | Use 404 instead of 403 when revealing the resource's existence is itself a leak |

## Implementation Notes

### Go

```go
// Middleware-based authorization
func RequireRole(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := auth.UserFromContext(r.Context())
            if user == nil {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
                return
            }
            if !slices.Contains(roles, user.Role) {
                http.Error(w, "forbidden", http.StatusForbidden)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Object-level check in handler
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
    user := auth.UserFromContext(r.Context())
    order, err := h.repo.GetByID(r.Context(), orderID)
    if err != nil { ... }

    if order.UserID != user.ID {
        http.Error(w, "not found", http.StatusNotFound) // 404, not 403
        return
    }
    // ...
}
```

### TypeScript (Express)

```typescript
import { Request, Response, NextFunction } from "express";

// Middleware-based authorization
function requireRole(...roles: string[]) {
  return (req: Request, res: Response, next: NextFunction) => {
    const user = req.user; // set by auth middleware
    if (!user) return res.status(401).json({ error: "unauthorized" });
    if (!roles.includes(user.role)) return res.status(403).json({ error: "forbidden" });
    next();
  };
}

// Usage
router.delete("/orders/:id", requireRole("admin", "editor"), async (req, res) => {
  const order = await orderRepo.getById(req.params.id);
  if (!order) return res.status(404).json({ error: "not found" });

  // Object-level check
  if (order.userId !== req.user!.id && req.user!.role !== "admin") {
    return res.status(404).json({ error: "not found" }); // 404, not 403
  }

  await orderRepo.delete(order.id);
  res.status(204).end();
});
```

For NestJS, use guards (`@UseGuards(RolesGuard)`) and custom decorators (`@Roles('admin')`). The pattern is the same: check role in middleware, check object ownership in the handler.

### Python (FastAPI)

```python
from fastapi import Depends, HTTPException

# Role-based dependency
def require_role(*roles: str):
    def checker(current_user: User = Depends(get_current_user)):
        if current_user.role not in roles:
            raise HTTPException(status_code=403, detail="Forbidden")
        return current_user
    return checker

@router.delete("/orders/{order_id}")
async def delete_order(
    order_id: UUID,
    current_user: User = Depends(require_role("admin", "editor")),
    repo: OrderRepo = Depends(get_order_repo),
):
    order = await repo.get_by_id(order_id)
    if not order:
        raise HTTPException(status_code=404)

    # Object-level authorization
    if order.user_id != current_user.id and current_user.role != "admin":
        raise HTTPException(status_code=404)  # hide existence

    await repo.delete(order_id)
```

## When to Bend the Rules

- **Public read-only endpoints** (product catalog, public profiles): No authorization needed, only rate limiting.
- **Internal-only services** behind a VPN or service mesh: Network-level access control may be sufficient, but still use service-to-service auth for defense in depth.
- **Prototypes and internal tools**: Start with simple role checks. Add object-level authorization before exposing to real users.
