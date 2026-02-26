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
| C-001 | Load config with defaults | Uses default values | ✓ |
| C-002 | Load config from YAML file | Reads all values from file | ✓ |
| C-003 | Env vars override file | Env takes precedence over file | ✓ |
| C-004 | Load config from env vars only | Uses env var values | ✓ |
| C-005 | Generate valid DSN | Correct PostgreSQL connection string | ✓ |

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

#### 2.2 Role Middleware
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| R-001 | No user in context | 401, unauthorized | ✓ |
| R-002 | ADMIN accessing ADMIN route | 200, allowed | ✓ |
| R-003 | OWNER accessing ADMIN route | 200, allowed (OWNER is superset) | ✓ |
| R-004 | ADMIN accessing OWNER-only route | 403, forbidden | ✓ |
| R-005 | Multiple roles allowed | 200, allowed if user has any | ✓ |

---

### 3. Integrations

#### 3.1 Shopify OAuth
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| O-001 | Generate OAuth URL | Valid URL with client_id, redirect_uri, state | ✓ |
| O-002 | Exchange code for token (success) | Access token returned | ✓ |
| O-003 | Exchange code for token (error) | Error returned | ✓ |
| O-004 | StartOAuth handler | 302 redirect to Shopify | ✓ |
| O-005 | Callback missing code | 400 bad request | ✓ |
| O-006 | Callback no user in context | 401 unauthorized | ✓ |
| O-007 | Callback success | 200, account created | ✓ |

#### 3.2 Encryption
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| E-001 | Encrypt and decrypt | Original data recovered | ✓ |
| E-002 | Same plaintext different ciphertext | Non-deterministic (random IV) | ✓ |
| E-003 | Invalid key length | Error returned | ✓ |
| E-004 | Invalid ciphertext | Error returned | ✓ |
| E-005 | Wrong key decryption | Error returned | ✓ |

#### 3.3 Manual Token (ADMIN only)
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| M-001 | Add token success | 201, token encrypted, masked in response | ✓ |
| M-002 | Add token missing token | 400, bad request | ✓ |
| M-003 | Add token missing partner_id | 400, bad request | ✓ |
| M-004 | Add token no user | 401, unauthorized | ✓ |
| M-005 | Add token updates existing | 200, account updated | ✓ |
| M-006 | Get token success | 200, masked token returned | ✓ |
| M-007 | Get token not found | 404, not found | ✓ |
| M-008 | Get token no user | 401, unauthorized | ✓ |
| M-009 | Revoke token success | 200, token deleted | ✓ |
| M-010 | Revoke token not found | 404, not found | ✓ |
| M-011 | Revoke token no user | 401, unauthorized | ✓ |
| M-012 | Mask token function | Correctly masks last 4 chars | ✓ |

#### 3.4 Shopify Partner Client
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| P-001 | Fetch apps success | Returns list of apps | ✓ |
| P-002 | Fetch apps GraphQL error | Returns error | ✓ |
| P-003 | Fetch apps HTTP error | Returns error | ✓ |
| P-004 | Fetch apps empty | Returns empty list | ✓ |

#### 3.5 App Management
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| AP-001 | Get available apps success | 200, apps from Partner API | ✓ |
| AP-002 | Get available apps no partner account | 404, not found | ✓ |
| AP-003 | Get available apps no user | 401, unauthorized | ✓ |
| AP-004 | Select app success | 201, app created | ✓ |
| AP-005 | Select app already exists | 409, conflict | ✓ |
| AP-006 | Select app missing fields | 400, bad request | ✓ |
| AP-007 | Select app no user | 401, unauthorized | ✓ |
| AP-008 | List apps success | 200, user's apps | ✓ |
| AP-009 | List apps no partner account | 404, not found | ✓ |
| AP-010 | List apps no user | 401, unauthorized | ✓ |

---

### 4. Domain

#### 4.1 LedgerService
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| LS-001 | Rebuild from transactions success | Subscriptions created, MRR calculated | ✓ |
| LS-002 | Separates RECURRING and USAGE | MRR = RECURRING only, Usage separate | ✓ |
| LS-003 | Computes expected renewal date | next_charge = last_charge + interval | ✓ |
| LS-004 | Classifies risk state | 31-60 days past due = ONE_CYCLE_MISSED | ✓ |
| LS-005 | Separate revenue by type | Returns recurring/usage arrays | ✓ |
| LS-006 | Deterministic rebuild | Same input → same output | ✓ |
| LS-007 | No transactions | Returns empty result | ✓ |
| LS-008 | Detects annual billing | 365-day pattern = ANNUAL | ✓ |

#### 4.2 Risk Engine
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| RE-001 | Active with future charge date | SAFE | ✓ |
| RE-002 | 0-30 days past due (grace period) | SAFE | ✓ |
| RE-003 | 31-60 days past due | ONE_CYCLE_MISSED | ✓ |
| RE-004 | 61-90 days past due | TWO_CYCLES_MISSED | ✓ |
| RE-005 | >90 days past due | CHURNED | ✓ |
| RE-006 | No expected charge date | SAFE (default) | ✓ |
| RE-007 | DaysPastDue nil charge date | Returns 0 | ✓ |
| RE-008 | DaysPastDue future charge date | Returns 0 | ✓ |
| RE-009 | DaysPastDue past charge date | Returns correct days | ✓ |
| RE-010 | RiskStateFromDaysPastDue boundary | Correct state at boundaries | ✓ |
| RE-011 | ClassifyAll batch operation | All subscriptions classified | ✓ |
| RE-012 | CalculateRiskSummary | Correct counts per state | ✓ |
| RE-013 | CalculateRevenueAtRisk | ONE_CYCLE + TWO_CYCLES MRR | ✓ |
| RE-014 | IsAtRisk helper | true for ONE/TWO CYCLE MISSED | ✓ |

#### 4.3 Metrics Engine
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| ME-001 | Calculate Active MRR | Sum of SAFE subscription MRR | ✓ |
| ME-002 | Calculate Active MRR annual | Annual / 12 = monthly | ✓ |
| ME-003 | Calculate Revenue at Risk | ONE_CYCLE + TWO_CYCLES MRR | ✓ |
| ME-004 | Calculate Usage Revenue | Sum of USAGE transactions | ✓ |
| ME-005 | Calculate Total Revenue | RECURRING + USAGE + ONE_TIME - REFUNDS | ✓ |
| ME-006 | Calculate Renewal Success Rate | SAFE / Total = decimal | ✓ |
| ME-007 | Renewal rate no subscriptions | Returns 0 | ✓ |
| ME-008 | Renewal rate all safe | Returns 1.0 | ✓ |
| ME-009 | Compute all metrics | Returns complete snapshot | ✓ |
| ME-010 | Compute metrics empty inputs | Returns zeros | ✓ |

#### 4.4 AI Insight Service
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| AI-001 | Generate insight Pro tier | 80-120 word brief returned | ✓ |
| AI-002 | Generate insight Free tier | ErrProTierRequired | ✓ |
| AI-003 | AI provider error | Error returned | ✓ |
| AI-004 | User not found | Error returned | ✓ |
| AI-005 | Build prompt | Contains key metrics | ✓ |

#### 4.3 Revenue Classification
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| RC-001 | AppSubscriptionSale | RECURRING | ✓ |
| RC-002 | AppUsageSale | USAGE | ✓ |
| RC-003 | AppOneTimeSale | ONE_TIME | ✓ |
| RC-004 | AppRefund | REFUND | ✓ |

---

### 5. Sync Engine

#### 5.1 SyncService
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| SS-001 | SyncApp success | Transactions fetched and stored | ✓ |
| SS-002 | SyncApp app not found | Error returned | ✓ |
| SS-003 | SyncApp fetch error | Error returned | ✓ |
| SS-004 | SyncApp no transactions | Success, 0 count | ✓ |
| SS-005 | SyncAllApps success | All apps synced, results returned | ✓ |

#### 5.2 SyncHandler
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| SH-001 | SyncAllApps success | 200, results array | ✓ |
| SH-002 | SyncAllApps no user | 401, unauthorized | ✓ |
| SH-003 | SyncAllApps no partner | 404, not found | ✓ |
| SH-004 | SyncApp success | 200, sync result | ✓ |
| SH-005 | SyncApp no user | 401, unauthorized | ✓ |
| SH-006 | SyncApp invalid ID | 400, bad request | ✓ |

#### 5.3 Transaction Sync (Future)
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| S-001 | First sync (empty DB) | All transactions imported | Pending |
| S-002 | Incremental sync | Only new transactions added | Pending |
| S-003 | Duplicate transaction | Ignored (idempotent) | Pending |
| S-004 | API rate limit hit | Retry with backoff | Pending |

---

### 6. API Endpoints (Future)

#### 6.1 Workspaces
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
