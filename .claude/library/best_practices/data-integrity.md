# Data Integrity

## Principle

Data corruption is worse than downtime. Design every data mutation so that the system remains consistent even when operations fail midway, run concurrently, or are retried. Use transactions, constraints, idempotency, and the expand-contract migration pattern to protect data at rest and in motion.

## Why

- **Corrupt data is permanent**: A crashed server restarts in minutes. Corrupt data requires investigation, manual fixes, and sometimes cannot be fully recovered. Every design decision should prefer "fail loudly and roll back" over "succeed partially and hope for the best."
- **Concurrency is the default**: Web applications serve multiple requests simultaneously. Any operation that reads, decides, and writes without protection will eventually produce wrong results under concurrent access.
- **Schema changes are the riskiest deploys**: More production incidents come from migrations than from application code. The expand-contract pattern exists because `ALTER TABLE DROP COLUMN` during a rolling deploy corrupts every request handled by the old code.

## Data Quality Dimensions

These eight dimensions (from DAMA-DMBOK) provide a vocabulary for discussing data issues:

| Dimension | Question to ask |
|-----------|----------------|
| **Accuracy** | Does the data reflect reality? Could it become stale? |
| **Completeness** | Are all required fields present? Can partial records be created? |
| **Consistency** | If the same fact exists in two places, can they diverge? |
| **Integrity** | Are relationships between records maintained? Can orphans be created? |
| **Reasonability** | Do values fall within plausible ranges? Are there missing bounds checks? |
| **Timeliness** | Is the data current enough for its purpose? Are caches expired correctly? |
| **Uniqueness** | Can duplicate records be created? Are unique constraints in place? |
| **Validity** | Do values conform to their expected type, format, and domain? |

## Core Rules

### 1. Use Transactions for Multi-Step Mutations

When an operation modifies multiple records, wrap all modifications in a transaction. If any step fails, all changes roll back.

```
BAD:  Create order → Deduct inventory → Charge payment
      (if payment fails, inventory is wrong)

GOOD: BEGIN → Create order → Deduct inventory → Charge payment → COMMIT
      (if payment fails, everything rolls back)
```

For operations spanning multiple services (where database transactions aren't possible), use the saga pattern: each step has a compensating action that undoes it on failure.

### 2. Enforce Constraints in the Database

Application-level validation can be bypassed (bugs, direct DB access, migrations). Database constraints are the last line of defense.

| Constraint | Protects against |
|-----------|-----------------|
| `NOT NULL` | Missing required data |
| `UNIQUE` | Duplicate records |
| `FOREIGN KEY` | Orphaned references |
| `CHECK` | Out-of-range values |
| `ON DELETE CASCADE` / `RESTRICT` | Dangling references when parent is deleted |

**Rule**: If a constraint matters, it must exist in the database, not only in application code.

### 3. Design for Idempotency

An operation is idempotent if performing it multiple times produces the same result as performing it once. This is essential because networks are unreliable — clients retry, webhooks fire twice, queues redeliver.

**Strategies**:
- **Natural idempotency**: `SET status = 'active'` is idempotent. `INCREMENT counter` is not.
- **Idempotency keys**: Client sends a unique key with the request. Server checks if it's been processed before.
- **Upsert**: `INSERT ... ON CONFLICT DO UPDATE` is idempotent by construction.

### 4. Prevent Race Conditions

Race conditions occur when concurrent operations read, decide, and write without coordination.

**Common patterns**:
- **Check-then-act**: "Is the seat available? Yes → Book it." Two requests check simultaneously, both see "available," both book → double booking.
- **Read-modify-write**: "Read balance → subtract 10 → write balance." Two concurrent subtractions each read the same starting balance → one subtraction is lost.

**Solutions**:
- **Optimistic locking**: Add a `version` column. Update with `WHERE version = ?`. If the update affects 0 rows, someone else modified the record — retry or fail.
- **Pessimistic locking**: `SELECT ... FOR UPDATE`. Blocks concurrent reads until the transaction commits.
- **Atomic operations**: `UPDATE accounts SET balance = balance - 10 WHERE balance >= 10`. No separate read step.
- **Unique constraints**: For "create if not exists" patterns, let the database enforce uniqueness rather than checking first.

### 5. Use Expand-Contract Migrations

Never make breaking schema changes in a single migration. Use four phases:

```
Phase 1 - EXPAND: Add new columns/tables with defaults. Never drop or rename.
Phase 2 - MIGRATE DATA: Backfill new structures from old.
Phase 3 - DEPLOY NEW CODE: Application reads/writes new structures.
Phase 4 - CONTRACT: Drop old columns/tables (separate migration, after all instances run new code).
```

**Rules**:
- Never add `NOT NULL` without a default value
- Never drop or rename a column in the same migration that adds its replacement
- Every migration must have a tested rollback script
- Rollback scripts must be idempotent (safe to run multiple times)

This pattern ensures zero-downtime deploys: during rollout, old and new code coexist, and both can read/write the database correctly.

### 6. Protect Referential Integrity

When deleting a parent record, handle its children explicitly:

| Strategy | When to use |
|----------|-------------|
| `CASCADE` | Children have no meaning without parent (order items when order is deleted) |
| `RESTRICT` | Children must be dealt with first (don't delete a user who has active orders) |
| `SET NULL` | Relationship is optional (set `assigned_to = NULL` when an employee leaves) |
| Soft delete | Never actually delete — set `deleted_at` timestamp, filter in queries |

**Never rely on application code alone** to maintain referential integrity. Use `ON DELETE CASCADE` or `ON DELETE RESTRICT` in the database.

### 7. Validate at System Boundaries

Validate data when it enters the system (API requests, file imports, message queue consumption) and when it leaves (API responses, exports). Internal code operating on already-validated data can trust the types.

This intersects with the defense-in-depth validation best practice. For data integrity specifically: validate that referenced records exist before creating relationships, and validate that state transitions are legal (e.g. an order can go from "pending" to "paid" but not from "cancelled" to "paid").

## Implementation Notes

### Go

```go
// Transaction wrapper
func (r *OrderRepo) CreateWithItems(ctx context.Context, order *Order, items []Item) error {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin tx: %w", err)
    }
    defer tx.Rollback() // no-op if committed

    if err := r.insertOrder(ctx, tx, order); err != nil {
        return fmt.Errorf("insert order: %w", err)
    }
    for _, item := range items {
        if err := r.insertItem(ctx, tx, &item); err != nil {
            return fmt.Errorf("insert item %s: %w", item.ID, err)
        }
    }

    return tx.Commit()
}

// Optimistic locking
func (r *AccountRepo) Debit(ctx context.Context, id uuid.UUID, amount int64, version int) error {
    result, err := r.db.ExecContext(ctx,
        `UPDATE accounts SET balance = balance - $1, version = version + 1
         WHERE id = $2 AND version = $3 AND balance >= $1`,
        amount, id, version,
    )
    if err != nil {
        return fmt.Errorf("debit: %w", err)
    }
    rows, _ := result.RowsAffected()
    if rows == 0 {
        return ErrConcurrentModification
    }
    return nil
}
```

### TypeScript

```typescript
// Transaction with Prisma
async function createWithItems(order: OrderInput, items: ItemInput[]): Promise<Order> {
  return prisma.$transaction(async (tx) => {
    const created = await tx.order.create({ data: order });

    await tx.orderItem.createMany({
      data: items.map((item) => ({ ...item, orderId: created.id })),
    });

    return created;
  });
}

// Optimistic locking with Prisma (manual version check)
async function debit(accountId: string, amount: number, expectedVersion: number): Promise<void> {
  const result = await prisma.account.updateMany({
    where: {
      id: accountId,
      version: expectedVersion,
      balance: { gte: amount },
    },
    data: {
      balance: { decrement: amount },
      version: { increment: 1 },
    },
  });

  if (result.count === 0) {
    throw new ConcurrentModificationError(accountId);
  }
}

// Transaction with Drizzle
import { db } from "./db";

await db.transaction(async (tx) => {
  await tx.insert(orders).values(order);
  await tx.insert(orderItems).values(items);
});
```

### Python

```python
# Transaction with SQLAlchemy async session
async def create_with_items(self, order: Order, items: list[OrderItem]) -> None:
    async with self.session.begin():  # auto-commits on success, rolls back on exception
        self.session.add(order)
        self.session.add_all(items)
        await self.session.flush()  # validate constraints before commit

# Optimistic locking with SQLAlchemy
class Account(Base):
    __tablename__ = "accounts"
    id: Mapped[UUID] = mapped_column(primary_key=True)
    balance: Mapped[int]
    version: Mapped[int] = mapped_column(default=0)

    __mapper_args__ = {"version_id_col": version}  # SQLAlchemy handles optimistic locking
```

## When to Bend the Rules

- **Analytics/logging data**: Eventual consistency is usually fine. Don't wrap analytics inserts in business transactions.
- **Append-only data** (event logs, audit trails): Immutable by design — many integrity concerns don't apply.
- **Prototyping**: Skip optimistic locking and expand-contract migrations. Add them before going to production.
- **Single-writer systems**: If only one process writes to a table, race conditions are structurally impossible. But verify this assumption is documented and enforced.
