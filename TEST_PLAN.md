# Test Plan – LedgerGuard

## Testing Strategy

### Unit Tests
- Test individual functions and methods in isolation
- Mock external dependencies (database, APIs)
- Located alongside source files (`*_test.go`)

### Integration Tests
- Test component interactions
- Use test database or in-memory implementations
- Located in `*_integration_test.go` files

### End-to-End Tests
- Test full API flows
- Run against test environment
- Located in `e2e/` directory (future)

---

## Test Scenarios

### 1. Infrastructure

#### 1.1 Health Endpoint
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| H-001 | GET /health with healthy DB | 200, status: "ok", database: "connected" | ✓ |
| H-002 | GET /health with unhealthy DB | 503, status: "degraded", database: "disconnected" | ✓ |
| H-003 | GET /health with no DB configured | 200, status: "ok", database: "not configured" | ✓ |

#### 1.2 Configuration
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| C-001 | Load config with defaults | Uses default values | Pending |
| C-002 | Load config from env vars | Uses env var values | Pending |
| C-003 | Generate valid DSN | Correct PostgreSQL connection string | Pending |

---

### 2. Authentication

#### 2.1 Auth Middleware
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| A-001 | Missing Authorization header | 401, unauthorized | ✓ |
| A-002 | Invalid Authorization format | 401, unauthorized | ✓ |
| A-003 | Invalid/expired token | 401, unauthorized | ✓ |
| A-004 | Valid token, existing user | 200, user in context | ✓ |
| A-005 | Valid token, new user | 200, user auto-created with OWNER role | ✓ |
| A-006 | Valid token, user creation fails | 500, internal error | ✓ |

---

### 3. Domain (Future)

#### 3.1 Risk Engine
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| R-001 | Active subscription | SAFE | Pending |
| R-002 | 0-30 days past due | SAFE | Pending |
| R-003 | 31-60 days past due | ONE_CYCLE_MISSED | Pending |
| R-004 | 61-90 days past due | TWO_CYCLE_MISSED | Pending |
| R-005 | >90 days past due | CHURNED | Pending |

#### 3.2 Revenue Classification
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| RC-001 | AppSubscriptionSale | RECURRING | Pending |
| RC-002 | AppUsageSale | USAGE | Pending |
| RC-003 | AppOneTimeSale | ONE_TIME | Pending |
| RC-004 | AppRefund | REFUND | Pending |

---

### 4. Sync Engine (Future)

#### 4.1 Transaction Sync
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| S-001 | First sync (empty DB) | All transactions imported | Pending |
| S-002 | Incremental sync | Only new transactions added | Pending |
| S-003 | Duplicate transaction | Ignored (idempotent) | Pending |
| S-004 | API rate limit hit | Retry with backoff | Pending |

---

### 5. API Endpoints (Future)

#### 5.1 Workspaces
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| W-001 | Create workspace | 201, workspace created | Pending |
| W-002 | Get workspace (owner) | 200, workspace data | Pending |
| W-003 | Get workspace (no access) | 403, forbidden | Pending |

---

## Running Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run specific package
go test ./internal/delivery/http -v

# Run with race detection
go test ./... -race
```

---

## Coverage Goals

| Package | Target | Current |
|---------|--------|---------|
| internal/domain | 90% | - |
| internal/usecase | 80% | - |
| internal/delivery/http | 70% | - |
| internal/repository | 70% | - |
| Overall | 75% | - |
