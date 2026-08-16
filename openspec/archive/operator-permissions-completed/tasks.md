# Operator Permissions - Implementation Tasks

**Specification**: [`spec.md`](./spec.md)  
**Priority**: CRITICAL  
**Effort**: 2-3 days  
**Status**: ✅ COMPLETE

> Implementation completed. See the delivery summary below.

---

## Delivery Summary

Implemented `operator_permission_checks` audit table, a role→permission matrix,
`RequirePermission`/`RequireAnyPermission` middleware, and wired permission
enforcement into both dashboard handlers (`DashboardHandler` and
`DashboardAPIHandler`). Permission checks are logged asynchronously.

### Files
- `migrations/0007_operator_permissions.up.sql` / `.down.sql`
- `handler/permissions.go` + `handler/permissions_test.go`
- `internal/storage/permissions.go` + `internal/storage/permissions_test.go`
- Modified: `handler/dashboard.go`, `handler/dashboard_api.go`, `main.go`,
  `handler/dashboard_test.go`

### Verification
- `go build ./...` passes
- `go test ./...` passes (handler + storage suites green)

---

## Task Checklist

### Phase 1: Database Migration

- [x] Create `migrations/0007_operator_permissions.up.sql`
  - [x] Create `operator_permission_checks` table
  - [x] Add indexes
- [x] Create `migrations/0007_operator_permissions.down.sql`
- [x] Role CHECK constraint verified as already present (added in 0005)

### Phase 2: Permission System Core

- [ ] Create `handler/permissions.go`
  - [ ] Define `Permission` type
  - [ ] Define all permission constants
  - [ ] Create `RolePermissions` map
  - [ ] Implement `HasPermission()` function
- [ ] Add permission tests
  - [ ] Test all roles
  - [ ] Test all permissions
  - [ ] Test edge cases

### Phase 3: Permission Middleware

- [ ] Implement `RequirePermission()` middleware
  - [ ] Extract operator from context
  - [ ] Check permission
  - [ ] Log permission check
  - [ ] Return 403 on denial
- [ ] Implement `RequireAnyPermission()` middleware
- [ ] Implement helper functions
  - [ ] `logPermissionCheck()`
  - [ ] `extractResource()`
  - [ ] `extractResourceID()`
  - [ ] `getClientIP()`
- [ ] Test middleware
  - [ ] Test with admin user
  - [ ] Test with operator user
  - [ ] Test with viewer user
  - [ ] Test without authentication

### Phase 4: Storage Layer

- [ ] Create `internal/storage/permissions.go`
  - [ ] Implement `LogPermissionCheck()` method
  - [ ] Add helper functions `nullString()`, `nullUUID()`
- [ ] Test storage layer
  - [ ] Test successful insert
  - [ ] Test with null values
  - [ ] Test performance with concurrent inserts

### Phase 5: Update Handlers

- [ ] Review all routes in `handler/dashboard_api.go`
- [ ] Add permission middleware to operator routes
  - [ ] `GET /operators` → `PermViewOperators`
  - [ ] `POST /operators` → `PermCreateOperator`
  - [ ] `PUT /operators/:id` → `PermUpdateOperator`
  - [ ] `DELETE /operators/:id` → `PermDeleteOperator`
  - [ ] `POST /operators/:id/totp-reset` → `PermResetTotp`
- [ ] Add permission middleware to audit routes
  - [ ] `GET /audit-logs` → `PermViewAuditLogs`
  - [ ] `GET /permission-checks` → `PermViewPermissionChecks`
- [ ] Add permission middleware to bot rules routes
  - [ ] `GET /bot-rules` → `PermViewBotRules` or `PermManageBotRules`
  - [ ] `POST /bot-rules/activate` → `PermManageBotRules`
- [ ] Add permission middleware to conversation routes
  - [ ] `GET /conversations` → `PermViewConversations`
  - [ ] `POST /conversations/:id/assign` → `PermAssignConversation`
  - [ ] `POST /conversations/:id/close` → `PermCloseConversation`
  - [ ] `POST /conversations/:id/reopen` → `PermReopenConversation`
  - [ ] `POST /conversations/:id/note` → `PermAddInternalNote`
- [ ] Add permission middleware to message routes
  - [ ] `POST /conversations/:id/message` → `PermSendMessages`
- [ ] Add permission middleware to invitation routes
  - [ ] `GET /invitations` → `PermInviteOperators`
  - [ ] `POST /invitations/whatsapp` → `PermInviteOperators`
  - [ ] `POST /invitations/email` → `PermInviteOperators`
  - [ ] `DELETE /invitations/:id` → `PermRevokeInvitation`
- [ ] Add permission middleware to team routes
  - [ ] `GET /team` → `PermManageTeam`
- [ ] Add permission middleware to account routes
  - [ ] `GET /accounts` → `PermViewAccounts` or `PermManageAccounts`
  - [ ] `POST /accounts` → `PermManageAccounts`
- [ ] Test all routes with different roles

### Phase 6: Testing

- [ ] Create `handler/permissions_test.go`
  - [ ] Write `TestHasPermission()`
  - [ ] Write `TestRequirePermission()`
  - [ ] Write `TestPermissionCheckLogging()`
- [ ] Run all tests
  - [ ] Ensure all tests pass
  - [ ] Check code coverage
- [ ] Integration testing
  - [ ] Test with real database
  - [ ] Test concurrent requests
  - [ ] Test error scenarios

### Phase 7: Manual Testing

- [ ] Create test operators for each role
  - [ ] Admin user
  - [ ] Operator user
  - [ ] Viewer user
- [ ] Test as admin
  - [ ] Verify can create operators
  - [ ] Verify can view audit logs
  - [ ] Verify can manage bot rules
  - [ ] Verify can perform all actions
- [ ] Test as operator
  - [ ] Verify cannot create operators
  - [ ] Verify cannot view audit logs
  - [ ] Verify can view conversations
  - [ ] Verify can send messages
  - [ ] Verify can close tickets
- [ ] Test as viewer
  - [ ] Verify cannot perform any write actions
  - [ ] Verify can only view data
- [ ] Verify logging
  - [ ] Check `operator_permission_checks` table
  - [ ] Verify IP and user agent recorded
  - [ ] Verify both allowed and denied checks logged

### Phase 8: Documentation

- [ ] Update API documentation
  - [ ] Document permission requirements for each endpoint
  - [ ] Add permission matrix to README
- [ ] Update `AGENT.md` checklist
  - [ ] Mark "Add operator roles and permissions" as complete
- [ ] Add migration notes
  - [ ] Document in CHANGELOG if exists
  - [ ] Update deployment guide

### Phase 9: Code Review

- [ ] Self-review
  - [ ] Check code style
  - [ ] Verify all tests pass
  - [ ] Check for security issues
- [ ] Request team review
  - [ ] Address feedback
  - [ ] Update code as needed
- [ ] Final verification
  - [ ] Run full test suite
  - [ ] Verify no regressions
  - [ ] Check performance metrics

### Phase 10: Deployment

- [ ] Prepare deployment
  - [ ] Create migration plan
  - [ ] Prepare rollback plan
- [ ] Deploy to staging
  - [ ] Run migration
  - [ ] Verify functionality
  - [ ] Run smoke tests
- [ ] Deploy to production
  - [ ] Run migration
  - [ ] Monitor logs
  - [ ] Verify permission checks working
  - [ ] Monitor performance

---

## Definition of Done

- [ ] All code written and tested
- [ ] All tests passing
- [ ] Migration tested (up and down)
- [ ] Documentation updated
- [ ] Manual testing complete
- [ ] Code review approved
- [ ] Deployed to staging
- [ ] Deployed to production
- [ ] Monitoring in place

---

## Blockers

None identified.

---

## Dependencies

- None (this is a foundational security feature)

---

## Notes

- This is a **CRITICAL** security feature - do not skip testing
- Fail closed: if permission system fails, deny access
- All permission checks must be logged for audit purposes
- Performance impact should be minimal (in-memory checks, async logging)

---

**Start Date**: August 16, 2025  
**Target Completion**: 2-3 days from start  
**Actual Completion**: August 16, 2025 (completed)
