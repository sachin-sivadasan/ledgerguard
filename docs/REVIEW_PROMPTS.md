# Review Prompts

Reusable prompts for code reviews and audits.

---

## Production Readiness Review

Use this prompt to identify issues before deploying to production:

```
Search the codebase for potential issues that need fixing. Look for:

1. TODO/FIXME comments
2. Placeholder implementations (functions returning nil, empty, or hardcoded values)
3. Missing error handling
4. Unimplemented interface methods
5. Nil pointer risks
6. Missing database migrations vs schema definitions
7. Environment variables referenced but potentially missing
8. Any panics or log.Fatal in non-main code

Provide a prioritized list of issues found with file paths and line numbers.
```

### Priority Levels

| Priority | Description | Action |
|----------|-------------|--------|
| CRITICAL | Will cause crashes or security vulnerabilities | Fix before deploy |
| HIGH | Core functionality broken or incomplete | Fix before deploy |
| MEDIUM | Edge cases, error handling gaps | Fix soon after deploy |
| LOWER | Code quality, minor improvements | Backlog |

---

## Security Review

```
Search the codebase for security vulnerabilities:

1. SQL injection risks (string concatenation in queries)
2. Missing authentication/authorization checks
3. Hardcoded secrets or credentials
4. Missing input validation
5. CSRF vulnerabilities
6. Insecure token handling
7. Missing rate limiting
8. Sensitive data in logs
9. Missing TLS/encryption
10. OWASP Top 10 vulnerabilities

Provide findings with severity, file paths, and recommended fixes.
```

---

## Performance Review

```
Search the codebase for performance issues:

1. N+1 query patterns
2. Missing database indexes
3. Unbounded queries (no LIMIT)
4. Large in-memory operations
5. Missing caching opportunities
6. Blocking operations in request handlers
7. Missing connection pooling
8. Inefficient loops or algorithms

Provide findings with impact assessment and optimization suggestions.
```

---

## Test Coverage Review

```
Analyze test coverage gaps:

1. Untested public functions
2. Missing edge case tests
3. Missing error path tests
4. Integration tests needed
5. Mock implementations that don't match interfaces
6. Tests that don't assert meaningful outcomes

List files/functions that need additional test coverage.
```
