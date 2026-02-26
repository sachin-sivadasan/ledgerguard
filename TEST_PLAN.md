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

#### 4.5 Notification Service
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| NS-001 | Register device success | Token stored, preferences created | ✓ |
| NS-002 | Register device invalid platform | ErrInvalidPlatform | ✓ |
| NS-003 | Register device duplicate same user | No error, no duplicate | ✓ |
| NS-004 | Register device transfers to new user | Token moved to new user | ✓ |
| NS-005 | Unregister device success | Token removed | ✓ |
| NS-006 | Unregister device not found | ErrDeviceTokenNotFound | ✓ |
| NS-007 | Unregister other user's token | ErrDeviceTokenNotFound | ✓ |
| NS-008 | Send critical alert success | Notification sent to all devices | ✓ |
| NS-009 | Send critical alert disabled | No notification sent | ✓ |
| NS-010 | Send critical alert no devices | No error, no notification | ✓ |
| NS-011 | Send daily summary success | Summary sent to all devices | ✓ |
| NS-012 | Send daily summary disabled | No notification sent | ✓ |
| NS-013 | Get preferences existing | Returns user preferences | ✓ |
| NS-014 | Get preferences not found | Returns default preferences | ✓ |
| NS-015 | Update preferences success | Preferences updated | ✓ |

#### 4.7 Slack Notification Provider
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| SL-001 | Send slack message success | Message delivered, 200 OK | ✓ |
| SL-002 | Send slack empty webhook URL | ErrInvalidWebhookURL | ✓ |
| SL-003 | Send slack non-200 response | ErrSlackWebhookFailed | ✓ |
| SL-004 | Send slack invalid URL | Error returned | ✓ |
| SL-005 | Send slack timeout | Error returned | ✓ |
| SL-006 | Slack color constants valid | All hex codes valid | ✓ |
| SL-007 | Slack integration critical alert | Sends to Slack with danger color | ✓ |
| SL-008 | Slack integration no webhook | No Slack message sent | ✓ |
| SL-009 | Slack integration daily summary | Sends to Slack with info color | ✓ |
| SL-010 | Slack + Push both configured | Both receive notifications | ✓ |
| SL-011 | Slack fails, push continues | Push still sent despite Slack error | ✓ |

#### 4.8 Revenue Classification
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

---

## Frontend Tests (Flutter)

### F1. Unit Tests (Bloc)

#### F1.1 AuthBloc
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-001 | Initial state | AuthInitial | ✓ |
| FB-002 | AuthCheckRequested user logged in | [AuthLoading, Authenticated] | ✓ |
| FB-003 | AuthCheckRequested user not logged in | [AuthLoading, Unauthenticated] | ✓ |
| FB-004 | SignInWithEmail success | [AuthLoading, Authenticated] | ✓ |
| FB-005 | SignInWithEmail invalid credentials | [AuthLoading, AuthError] | ✓ |
| FB-006 | SignInWithEmail user not found | [AuthLoading, AuthError] | ✓ |
| FB-007 | SignInWithGoogle success | [AuthLoading, Authenticated] | ✓ |
| FB-008 | SignInWithGoogle cancelled | [AuthLoading, Unauthenticated] | ✓ |
| FB-009 | SignInWithGoogle failure | [AuthLoading, AuthError] | ✓ |
| FB-010 | SignOut success | [AuthLoading, Unauthenticated] | ✓ |
| FB-011 | SignOut failure | [AuthLoading, AuthError] | ✓ |

#### F1.2 OnboardingBloc
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-020 | Connect Shopify Partner | Navigate to app selection | Pending |
| FB-021 | Select app success | Emit app selected state | Pending |
| FB-022 | Select app failure | Emit error state | Pending |

#### F1.3 DashboardBloc
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-030 | Initial state | DashboardInitial | ✓ |
| FB-031 | Load metrics success | [Loading, Loaded] | ✓ |
| FB-032 | Load metrics failure | [Loading, Error] | ✓ |
| FB-033 | Load metrics empty | [Loading, Empty] | ✓ |
| FB-034 | Refresh metrics success | [Loaded(refreshing), Loaded] | ✓ |
| FB-035 | Refresh metrics failure | [Loaded(refreshing), Loaded] keeps data | ✓ |
| FB-036 | Refresh when not loaded | Triggers load | ✓ |
| FB-037 | Refresh returns empty | [Loaded(refreshing), Empty] | ✓ |

#### F1.4 SubscriptionBloc
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-040 | Load subscriptions | Emit list of subscriptions | Pending |
| FB-041 | Filter by risk state | Emit filtered list | Pending |
| FB-042 | View subscription detail | Emit detail state | Pending |

#### F1.5 PartnerIntegrationBloc
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-070 | Initial state | PartnerIntegrationInitial | ✓ |
| FB-071 | CheckStatus when connected | [Loading, Connected] | ✓ |
| FB-072 | CheckStatus when not connected | [Loading, NotConnected] | ✓ |
| FB-073 | CheckStatus failure | [Loading, Error] | ✓ |
| FB-074 | ConnectWithOAuth success | [Loading, Success] | ✓ |
| FB-075 | ConnectWithOAuth failure | [Loading, Error] | ✓ |
| FB-076 | SaveManualToken success | [Loading, Success] | ✓ |
| FB-077 | SaveManualToken invalid | [Loading, Error] | ✓ |
| FB-078 | Disconnect success | [Loading, NotConnected] | ✓ |
| FB-079 | Disconnect failure | [Loading, Error] | ✓ |

#### F1.6 AppSelectionBloc
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-080 | Initial state | AppSelectionInitial | ✓ |
| FB-081 | FetchApps success | [Loading, Loaded] | ✓ |
| FB-082 | FetchApps with previous selection | [Loading, Loaded(selectedApp)] | ✓ |
| FB-083 | FetchApps no apps | [Loading, Error] | ✓ |
| FB-084 | FetchApps failure | [Loading, Error] | ✓ |
| FB-085 | AppSelected updates state | Loaded with selectedApp | ✓ |
| FB-086 | AppSelected in wrong state | No change | ✓ |
| FB-087 | ConfirmSelection success | [Saving, Confirmed] | ✓ |
| FB-088 | ConfirmSelection failure | [Saving, Error] | ✓ |
| FB-089 | ConfirmSelection no selection | No change | ✓ |
| FB-090 | LoadSelectedApp exists | [Confirmed] | ✓ |
| FB-091 | LoadSelectedApp not exists | No change | ✓ |

---

### F2. UseCases

#### F2.1 Authentication
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FU-001 | GetCurrentUser | Returns user or null | Pending |
| FU-002 | SignInWithEmail | Returns user on success | Pending |
| FU-003 | SignOut | Clears user session | Pending |

#### F2.2 Metrics
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FU-010 | GetDashboardMetrics | Returns KPI snapshot | Pending |
| FU-011 | GetSubscriptions | Returns subscription list | Pending |
| FU-012 | GetSubscriptionDetail | Returns single subscription | Pending |

---

### F3. Repository Tests

#### F3.1 AuthRepository
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FR-001 | Sign in with Firebase | Returns user data | Pending |
| FR-002 | Token refresh | New token fetched | Pending |
| FR-003 | Sign out | Session cleared | Pending |

#### F3.2 MetricsRepository
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FR-010 | Fetch metrics from API | Returns parsed metrics | Pending |
| FR-011 | API error handling | Throws domain exception | Pending |

---

### F4. Widget Tests

#### F4.1 Core Widgets
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-001 | MetricCard renders | Shows value and label | Pending |
| FW-002 | RiskBadge colors | Correct color per state | Pending |
| FW-003 | SubscriptionTile | Shows subscription info | Pending |

#### F4.2 LoginPage
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-010 | Renders email and password fields | Two TextFormFields visible | ✓ |
| FW-011 | Renders login button | Sign In button visible | ✓ |
| FW-012 | Renders Google sign in button | Continue with Google visible | ✓ |
| FW-013 | Renders signup link | Don't have an account? + Sign Up | ✓ |
| FW-014 | Shows loading indicator | CircularProgressIndicator when AuthLoading | ✓ |
| FW-015 | Shows error message | Error text when AuthError | ✓ |
| FW-016 | Dispatches SignInWithEmail | Event added on button tap | ✓ |
| FW-017 | Dispatches SignInWithGoogle | Event added on Google button tap | ✓ |
| FW-018 | Disables buttons when loading | onPressed is null | ✓ |

#### F4.3 SignupPage
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-020 | Renders email and password fields | Two TextFormFields visible | ✓ |
| FW-021 | Renders create account button | Create Account button visible | ✓ |
| FW-022 | Renders Google sign up button | Continue with Google visible | ✓ |
| FW-023 | Renders login link | Already have an account? + Sign In | ✓ |
| FW-024 | Shows loading indicator | CircularProgressIndicator when AuthLoading | ✓ |
| FW-025 | Shows error message | Error text when AuthError | ✓ |
| FW-026 | Dispatches SignInWithGoogle | Event added on Google button tap | ✓ |
| FW-027 | Disables buttons when loading | onPressed is null | ✓ |

#### F4.4 RoleGuard Widget
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-030 | Shows child for owner when owner required | Child visible | ✓ |
| FW-031 | Hides child for admin when owner required | Child hidden | ✓ |
| FW-032 | Shows child for owner when admin required | Child visible | ✓ |
| FW-033 | Shows child for admin when admin required | Child visible | ✓ |
| FW-034 | Shows fallback when role not met | Fallback widget visible | ✓ |
| FW-035 | Shows nothing when role not loaded | SizedBox.shrink | ✓ |
| FW-036 | Shows loading when showLoading true | CircularProgressIndicator | ✓ |

#### F4.5 ProGuard Widget
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-040 | Shows child for Pro tier | Child visible | ✓ |
| FW-041 | Hides child for Starter tier | Child hidden | ✓ |
| FW-042 | Shows fallback for Starter tier | Fallback widget visible | ✓ |

#### F4.6 ManualIntegrationPage
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-050 | Shows content for owner | Form fields visible | ✓ |
| FW-051 | Shows content for admin | Form fields visible | ✓ |
| FW-052 | Shows loading while role loading | CircularProgressIndicator | ✓ |
| FW-053 | Shows access denied for non-admin | Access denied message | ✓ |

#### F4.7 PartnerIntegrationPage
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-060 | Renders page title | "Partner Integration" visible | ✓ |
| FW-061 | Renders OAuth connect button | "Connect Shopify Partner" visible | ✓ |
| FW-062 | Dispatches ConnectWithOAuth | Event added on button tap | ✓ |
| FW-063 | Shows Manual Token for admin | Form fields visible for admin | ✓ |
| FW-064 | Shows Manual Token for owner | Form fields visible for owner | ✓ |
| FW-065 | Hides Manual Token when role not loaded | Form hidden | ✓ |
| FW-066 | Dispatches SaveManualToken | Event added with form values | ✓ |
| FW-067 | Validates Partner ID required | Error message shown | ✓ |
| FW-068 | Validates API Token required | Error message shown | ✓ |
| FW-069 | Shows loading indicator | CircularProgressIndicator when loading | ✓ |
| FW-070 | Shows connected state | Partner ID and Disconnect visible | ✓ |
| FW-071 | Dispatches Disconnect | Event added on button tap | ✓ |
| FW-072 | Shows error message | Error text when error state | ✓ |
| FW-073 | Shows success state | Connected card visible | ✓ |
| FW-074 | Hides buttons when loading | Connect buttons not visible | ✓ |
| FW-075 | Checks status on init | Event added on page load | ✓ |

#### F4.8 AppSelectionPage
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-080 | Renders page title | "Select App" visible | ✓ |
| FW-081 | Fetches apps on init | FetchAppsRequested dispatched | ✓ |
| FW-082 | Shows loading indicator | CircularProgressIndicator when loading | ✓ |
| FW-083 | Shows error message | Error text when error state | ✓ |
| FW-084 | Dispatches FetchApps on retry | Event added on retry tap | ✓ |
| FW-085 | Shows list of apps | App names visible | ✓ |
| FW-086 | Shows app descriptions | Description text visible | ✓ |
| FW-087 | Shows install counts | Install count visible | ✓ |
| FW-088 | Dispatches AppSelected on tap | Event added with app | ✓ |
| FW-089 | Shows Confirm button when selected | "Confirm Selection" visible | ✓ |
| FW-090 | Hides Confirm button when not selected | Button not visible | ✓ |
| FW-091 | Dispatches ConfirmSelection | Event added on confirm tap | ✓ |
| FW-092 | Shows saving indicator | "Saving..." text visible | ✓ |
| FW-093 | Disables selection when saving | Taps don't dispatch events | ✓ |
| FW-094 | Shows selected app indicator | Check icon visible | ✓ |

#### F4.9 DashboardPage
| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-100 | Renders page title | "Executive Dashboard" visible | ✓ |
| FW-101 | Fetches metrics on init | LoadDashboardRequested dispatched | ✓ |
| FW-102 | Shows loading indicator | CircularProgressIndicator when loading | ✓ |
| FW-103 | Shows error message | Error text when error state | ✓ |
| FW-104 | Dispatches LoadDashboard on retry | Event added on retry tap | ✓ |
| FW-105 | Empty state shows message | "No Metrics Yet" visible | ✓ |
| FW-106 | Empty state shows sync button | "Sync Data" button visible | ✓ |
| FW-107 | Empty state dispatches refresh | RefreshDashboardRequested on tap | ✓ |
| FW-108 | Empty state custom message | Custom message visible | ✓ |
| FW-109 | Displays Renewal Success Rate | Rate percentage visible | ✓ |
| FW-106 | Displays Active MRR | MRR value visible | ✓ |
| FW-107 | Displays Revenue at Risk | Risk value visible | ✓ |
| FW-108 | Displays Churned metrics | Churned value and count visible | ✓ |
| FW-109 | Displays Usage Revenue | Usage value visible | ✓ |
| FW-110 | Displays Total Revenue | Total revenue value visible | ✓ |
| FW-111 | Displays Revenue Mix chart | Recurring/Usage/One-time visible | ✓ |
| FW-112 | Displays Risk Distribution chart | Safe/At Risk/Critical visible | ✓ |
| FW-113 | Displays risk counts | Count numbers visible | ✓ |
| FW-114 | Shows refresh button | Refresh icon in app bar | ✓ |
| FW-115 | Dispatches Refresh on tap | RefreshDashboardRequested dispatched | ✓ |
| FW-116 | Shows loading when refreshing | Progress indicator when refreshing | ✓ |
| FW-117 | Disables refresh when refreshing | Refresh icon hidden | ✓ |
| FW-118 | Displays Primary KPIs header | Section header visible | ✓ |
| FW-119 | Displays Revenue & Risk header | Section header visible | ✓ |
| FW-120 | Shows settings button | Settings icon in app bar | ✓ |

#### F4.10 PreferencesBloc

| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-070 | Initial state | PreferencesInitial | ✓ |
| FB-071 | Load preferences success | [Loading, Loaded] | ✓ |
| FB-072 | Load preferences failure | [Loading, Error] | ✓ |
| FB-073 | Add primary KPI | KPI added to preferences | ✓ |
| FB-074 | Add duplicate KPI | KPI not duplicated | ✓ |
| FB-075 | Max 4 KPIs enforced | No more than 4 KPIs | ✓ |
| FB-076 | Remove primary KPI | KPI removed | ✓ |
| FB-077 | Remove non-existent KPI | No change | ✓ |
| FB-078 | Reorder KPIs | KPIs reordered correctly | ✓ |
| FB-079 | Toggle secondary widget (disable) | Widget disabled | ✓ |
| FB-080 | Toggle secondary widget (enable) | Widget enabled | ✓ |
| FB-081 | Save preferences success | [Saving, Saved, Loaded] | ✓ |
| FB-082 | Save preferences failure | [Saving, Error, Loaded] | ✓ |
| FB-083 | Reset preferences | Defaults restored | ✓ |

#### F4.11 DashboardConfigDialog

| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-130 | Shows loading state | CircularProgressIndicator visible | ✓ |
| FW-131 | Shows error with retry | Error message and retry button | ✓ |
| FW-132 | Shows preferences when loaded | Configuration UI visible | ✓ |
| FW-133 | Displays all primary KPIs | KPI list visible | ✓ |
| FW-134 | Displays all secondary widgets | Widget toggles visible | ✓ |
| FW-135 | Shows unsaved changes indicator | "Unsaved changes" text | ✓ |
| FW-136 | Remove KPI dispatches event | RemovePrimaryKpiRequested | ✓ |
| FW-137 | Toggle widget dispatches event | ToggleSecondaryWidgetRequested | ✓ |
| FW-138 | Save button disabled when no changes | Button disabled | ✓ |
| FW-139 | Save button enabled when changes | Button enabled | ✓ |
| FW-140 | Reset button dispatches event | ResetPreferencesRequested | ✓ |
| FW-141 | Shows KPI count | "(N/4)" format | ✓ |
| FW-142 | Shows Add KPI when under max | "Add KPI" button visible | ✓ |
| FW-143 | Hides Add KPI when at max | Button hidden | ✓ |
| FW-144 | Shows saving indicator | CircularProgressIndicator in button | ✓ |

#### F4.12 RiskBloc

| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-090 | Initial state | RiskInitial | ✓ |
| FB-091 | Load success | [Loading, Loaded] | ✓ |
| FB-092 | Load returns null | [Loading, Empty] | ✓ |
| FB-093 | Load returns empty data | [Loading, Empty] | ✓ |
| FB-094 | Load fails | [Loading, Error] | ✓ |
| FB-095 | Refresh success | [Loaded(refreshing), Loaded] | ✓ |
| FB-096 | Refresh fails | [Loaded(refreshing), Loaded] with original data | ✓ |
| FB-097 | Refresh when not loaded | Triggers load | ✓ |

#### F4.13 RiskSummary Entity

| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FE-010 | totalSubscriptions calculation | Sum of all counts | ✓ |
| FE-011 | percentFor calculation | Correct percentage | ✓ |
| FE-012 | atRiskCount calculation | oneCycle + twoCycles | ✓ |
| FE-013 | formattedRevenueAtRisk | Currency format | ✓ |
| FE-014 | hasData check | true when total > 0 | ✓ |

#### F4.14 RiskBreakdownPage

| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FW-150 | Shows loading state | CircularProgressIndicator visible | ✓ |
| FW-151 | Shows empty state | "No Risk Data" message | ✓ |
| FW-152 | Shows error with retry | Error message and retry button | ✓ |
| FW-153 | Shows app bar title | "Risk Breakdown" visible | ✓ |
| FW-154 | Shows total subscriptions | Count displayed | ✓ |
| FW-155 | Shows revenue at risk | Currency value displayed | ✓ |
| FW-156 | Shows distribution section | Section header visible | ✓ |
| FW-157 | Shows breakdown by state | Section header visible | ✓ |
| FW-158 | Shows all risk states | SAFE, ONE_CYCLE_MISSED, etc. | ✓ |
| FW-159 | Shows legend items | Color legend visible | ✓ |
| FW-160 | Shows refresh button | Refresh icon in app bar | ✓ |
| FW-161 | Dispatches refresh on tap | RefreshRiskSummaryRequested | ✓ |
| FW-162 | Shows loading when refreshing | Progress indicator in app bar | ✓ |
| FW-163 | Shows pie chart | CustomPaint widget visible | ✓ |
| FW-164 | Shows risk state descriptions | Description text visible | ✓ |

---

### F5. RoleBloc Tests

| ID | Scenario | Expected Result | Status |
|----|----------|-----------------|--------|
| FB-050 | Initial state | RoleInitial | ✓ |
| FB-051 | FetchRole success (owner) | [RoleLoading, RoleLoaded] | ✓ |
| FB-052 | FetchRole success (admin) | [RoleLoading, RoleLoaded] | ✓ |
| FB-053 | FetchRole profile not found | [RoleLoading, RoleError] | ✓ |
| FB-054 | FetchRole unauthorized | [RoleLoading, RoleError] | ✓ |
| FB-055 | ClearRole clears cache | [RoleInitial] | ✓ |
| FB-056 | RoleLoaded.isOwner for owner | true | ✓ |
| FB-057 | RoleLoaded.isOwner for admin | false | ✓ |
| FB-058 | RoleLoaded.isPro for pro tier | true | ✓ |
| FB-059 | RoleLoaded.isPro for starter tier | false | ✓ |
| FB-060 | RoleLoaded.hasRole checks permission | Correct permission check | ✓ |

---

### Running Frontend Tests

```bash
# Run all frontend tests
cd frontend/app && flutter test

# Run with coverage
flutter test --coverage

# Run specific test file
flutter test test/blocs/auth_bloc_test.dart

# Watch mode
flutter test --watch
```

---

### Frontend Coverage Goals

| Layer | Target | Current |
|-------|--------|---------|
| Blocs | 90% | - |
| UseCases | 90% | - |
| Repositories | 80% | - |
| Widgets | 70% | - |
| Overall | 80% | - |
