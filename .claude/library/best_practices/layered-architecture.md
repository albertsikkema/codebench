# Layered Architecture

## Principle

Separate application code into distinct layers with clear responsibilities and strict dependency direction. Each layer only depends on the layer below it — never sideways or upward.

```
Handler / API Layer
    ↓
Service / Business Logic Layer
    ↓
Repository / Data Access Layer
    ↓
Domain Model Layer
```

## Why

- **Testability**: Each layer can be tested in isolation by mocking the layer below.
- **Replaceability**: Swap a database, framework, or transport protocol without rewriting business logic.
- **Readability**: A new developer knows where to look — HTTP concerns in handlers, rules in services, queries in repositories.
- **Enforced boundaries**: Prevents the common failure mode where database queries, business rules, and HTTP responses are tangled in a single function.

## Layer Responsibilities

### 1. Handler / API Layer

Owns the transport protocol. Accepts requests, validates input shape, delegates to services, and formats responses.

**Does:**
- Parse and validate incoming requests (deserialization, schema validation)
- Map service results to response formats (status codes, headers, serialization)
- Inject dependencies into services
- Document the external API (OpenAPI, gRPC proto, CLI help text)

**Does not:**
- Contain business rules or conditional logic beyond input validation
- Access the database or any external system directly
- Decide what data to return based on domain rules

### 2. Service / Business Logic Layer

Owns domain rules and workflows. Orchestrates repositories and external services.

**Does:**
- Implement business rules ("a user can only have 5 active sessions")
- Coordinate multiple repositories within a single operation
- Manage transaction boundaries (begin, commit, rollback)
- Raise domain-specific errors
- Log structured events with business context

**Does not:**
- Know about HTTP, gRPC, or any transport protocol
- Construct database queries
- Handle serialization or response formatting

### 3. Repository / Data Access Layer

Owns all interaction with persistent storage. Exposes domain-oriented query methods, hides query implementation.

**Does:**
- Encapsulate all database queries behind named methods (`GetUserByEmail`, `ListActiveOrders`)
- Handle query optimisation (eager loading, pagination, indexing hints)
- Filter by authorisation context (e.g. tenant ID, user ID) when appropriate
- Map between database representations and domain types

**Does not:**
- Contain business logic or conditional workflows
- Know about the transport layer
- Expose raw query builders or ORM internals to callers

### 4. Domain Model Layer

Owns the shape of your data. Defines entities, value objects, and their constraints.

**Does:**
- Define data structures and their fields
- Express database-level constraints (unique, not null, foreign keys)
- Define relationships between entities
- Carry type information used by all other layers

**Does not:**
- Perform operations or side effects
- Depend on any other layer

## Dependency Direction

The strict rule: **dependencies point downward only**.

```
Handler  →  Service  →  Repository  →  Model
   ✗           ✗            ✗
   ↑           ↑            ↑
 never       never        never
```

- A handler may call a service. A service may call a repository. A repository may use a model.
- A repository must never import from a handler. A service must never import from a handler.
- A model must never import from any other layer.

Violations of this rule create circular dependencies and make testing progressively harder.

### Horizontal Dependencies

**Services can call other services.** An `OrderService` that needs to check inventory and charge a payment legitimately depends on `InventoryService` and `PaymentService`. This is normal orchestration.

```
OrderService  →  InventoryService
              →  PaymentService
```

**Rules for horizontal service dependencies:**
- Keep the dependency graph acyclic. If A depends on B and B depends on A, extract the shared logic into a third service or push coordination up to the caller.
- Inject service dependencies the same way as repositories — through constructor arguments or dependency injection, never by direct instantiation.
- If a service needs only one method from another service, consider whether it actually needs the repository directly instead. Don't route through a service layer just for a simple data lookup.

**Repositories can be used by multiple services.** A `UserRepository` might be used by `AuthService`, `ProfileService`, and `AdminService`. This is expected — repositories are shared data access, not owned by a single service.

```
AuthService     →  UserRepository
ProfileService  →  UserRepository
AdminService    →  UserRepository
```

**Repositories must never call services.** This is an upward dependency and violates the core rule. If a repository seems to need business logic, that logic belongs in the service that called the repository.

**Repositories should not call other repositories.** If a repository needs data from another table, that's usually a join in a single query, not a call to another repository. Cross-repository coordination belongs in the service layer, where transaction boundaries are managed.

## Common Mistakes

### Business logic in the handler

```
BAD:  Handler checks "if user.role == admin" before querying the database
GOOD: Handler calls service.GetResource(userID, resourceID)
      Service checks authorisation, raises error if denied
```

### Database queries in the service

```
BAD:  Service constructs SQL or ORM queries directly
GOOD: Service calls repository.FindByStatus("active")
      Repository owns the query implementation
```

### Transport concerns in the service

```
BAD:  Service raises HTTPException or sets status codes
GOOD: Service raises DomainError("insufficient balance")
      Handler catches DomainError and maps to HTTP 422
```

### God repository with raw query exposure

```
BAD:  Repository exposes a generic Query(filter) method that accepts raw WHERE clauses
GOOD: Repository exposes FindActiveByUser(userID) — named, typed, tested
```

## Implementation Notes

### Go

The standard Go project layout maps naturally:

| Layer | Package | Typical types |
|-------|---------|---------------|
| Handler | `handler/` or `api/` | HTTP handlers, middleware, request/response structs |
| Service | `service/` | Business logic structs with repository interfaces as fields |
| Repository | `store/` or `repo/` | Structs wrapping `*sql.DB` or `*sqlx.DB` |
| Model | `model/` or `domain/` | Plain structs, often shared across layers |

Define repository interfaces in the service package so services depend on abstractions, not implementations. This makes testing with mock repositories straightforward.

```go
// service/user.go
type UserRepository interface {
    GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
    Save(ctx context.Context, user *model.User) error
}

type UserService struct {
    repo UserRepository
}
```

### TypeScript (Express / NestJS)

| Layer | Module | Typical types |
|-------|--------|---------------|
| Handler | `controllers/` or `routes/` | Express route handlers, NestJS controllers with decorators |
| Service | `services/` | Classes or functions with injected repositories |
| Repository | `repositories/` | Classes wrapping Prisma, Drizzle, or TypeORM |
| Model | `models/` or `entities/` | Prisma schema, TypeORM entities, or plain TypeScript types |

Use constructor injection (NestJS) or factory functions (Express) for dependency injection. Define repository interfaces so services depend on abstractions.

```typescript
// services/user.service.ts
export interface UserRepository {
  getById(id: string): Promise<User | null>;
  save(user: User): Promise<User>;
}

export class UserService {
  constructor(private readonly repo: UserRepository) {}

  async getById(id: string): Promise<User> {
    const user = await this.repo.getById(id);
    if (!user) throw new UserNotFoundError(id);
    return user;
  }
}
```

### Python (FastAPI)

| Layer | Module | Typical types |
|-------|--------|---------------|
| Handler | `routers/` | FastAPI router functions with `Depends()` |
| Service | `services/` | Classes or functions with injected repositories |
| Repository | `repositories/` | Classes wrapping `AsyncSession` |
| Model | `models/` | SQLAlchemy ORM models |

Use FastAPI's `Depends()` for dependency injection. Define Pydantic schemas separate from ORM models.

```python
# routers/user.py
@router.get("/users/{user_id}")
async def get_user(
    user_id: UUID,
    service: UserService = Depends(get_user_service),
) -> UserResponse:
    user = await service.get_by_id(user_id)
    return UserResponse.model_validate(user)
```

## When to Bend the Rules

- **Small scripts or CLIs**: A 200-line CLI tool does not need four layers. Start with two (handler + data access) and add services when business logic appears.
- **Prototypes**: Collapsing layers is fine when exploring. Separate them before the code goes to production.
- **Performance-critical paths**: Occasionally a service needs a specialised query that doesn't fit the repository's abstraction. Prefer adding a named repository method over bypassing the layer.

The goal is not ceremony — it is knowing where every piece of logic lives and being able to test it independently.
