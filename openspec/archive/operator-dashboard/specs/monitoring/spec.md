# Error logging and monitoring requirements for operator dashboard

## ADDED Requirements

### Requirement: Client errors are reported

Uncaught JavaScript errors MUST be reported to a monitoring service.

#### Scenario: Runtime error

- **WHEN** a JavaScript error occurs in production
- **THEN** it MUST be captured and sent to the error monitoring backend
- **AND** include stack trace, user agent, route, and Redux/state snapshot (sanitized)

### Requirement: Failed API requests are logged

Network failures and API errors MUST be logged with context.

#### Scenario: API 500 error

- **WHEN** an API request fails with 500
- **THEN** the error MUST be logged with endpoint, status, timestamp
- **AND** not include sensitive request/response bodies

#### Scenario: Network timeout

- **WHEN** a request times out
- **THEN** it MUST be logged as a network error with duration

### Requirement: Performance metrics are collected

Core Web Vitals and custom performance metrics MUST be tracked.

#### Scenario: Page load metrics

- **WHEN** a page loads
- **THEN** metrics like FCP, LCP, TTI MUST be recorded
- **AND** sent to the analytics backend

#### Scenario: API latency

- **WHEN** an API request completes
- **THEN** duration MUST be recorded
- **AND** tagged by endpoint and success/failure

### Requirement: Logs respect privacy

All logging MUST avoid PII and sensitive data.

#### Scenario: Sanitized logging

- **WHEN** logging errors or metrics
- **THEN** message content, phone numbers, and auth tokens MUST be redacted
- **AND** only log metadata (conversation ID, not content)
