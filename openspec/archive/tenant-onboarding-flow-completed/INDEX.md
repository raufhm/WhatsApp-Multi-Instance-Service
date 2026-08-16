# 📚 TOTP-Based Onboarding Specification - Complete Index

**WhatsApp Multi-Instance Service - Passwordless Authentication**

---

## 🎯 Quick Navigation

### For Product Managers
- **Start here**: [`README.md`](README.md) - Complete overview and business value
- **User flows**: [`README.md`](README.md#key-features) - How users experience TOTP
- **Benefits**: [`proposal.md`](proposal.md) - Why TOTP over passwords
- **Timeline**: [`IMPLEMENTATION-STATUS.md`](IMPLEMENTATION-STATUS.md) - 6-8 week roadmap

### For Developers
- **Quick start**: [`QUICKSTART.md`](QUICKSTART.md) - Code examples and setup
- **UI implementation**: [`UI-MIGRATION-TANSTACK.md`](UI-MIGRATION-TANSTACK.md) - Frontend with TanStack
- **Backend guide**: [`TOTP-MIGRATION.md`](TOTP-MIGRATION.md) - API and database migration
- **Database schema**: [`design.md`](design.md) - Complete schema reference

### For Architects
- **Architecture**: [`design.md`](design.md) - System design and data flow
- **Security**: [`specs/totp-authentication/spec.md`](specs/totp-authentication/spec.md) - Security measures
- **Recovery flows**: [`specs/totp-reset-recovery/spec.md`](specs/totp-reset-recovery/spec.md) - Backup codes, admin reset

### For QA/Testing
- **API specs**: [`specs/`](specs/) - All endpoint requirements
- **Test scenarios**: [`UI-MIGRATION-TANSTACK.md`](UI-MIGRATION-TANSTACK.md#testing-strategy) - E2E test examples
- **Acceptance criteria**: Each spec file includes verification steps

---

## 📁 Document Structure

```
openspec/changes/tenant-onboarding-flow/
├── INDEX.md                          # You are here (navigation guide)
├── README.md                         # Complete overview (14KB)
├── proposal.md                       # High-level proposal (2KB)
├── design.md                         # Database schema & architecture (~12KB)
├── TOTP-MIGRATION.md                 # Backend implementation guide (8KB)
├── QUICKSTART.md                     # Developer quick reference (8KB)
├── UI-MIGRATION-TANSTACK.md          # Frontend implementation guide (42KB)
├── IMPLEMENTATION-STATUS.md          # Status tracker & checklist (8KB)
└── specs/
    ├── signin/spec.md                # TOTP login requirements (17KB)
    ├── signup/spec.md                # TOTP signup requirements (17KB)
    ├── totp-authentication/
    │   └── spec.md                   # Core TOTP implementation (21KB)
    ├── totp-reset-recovery/
    │   └── spec.md                   # Recovery flows (13KB)
    ├── whatsapp-invitations/
    │   └── spec.md                   # WhatsApp invitations (17KB)
    └── email-verification/
        └── spec.md                   # Email verification (legacy, admin-only)
```

---

## 📊 Document Summary

### Core Documents

| Document | Purpose | Audience | Size |
|----------|---------|----------|------|
| [README.md](README.md) | Complete overview with business value, user flows, timeline | PM, Dev, QA | 14KB |
| [proposal.md](proposal.md) | Why TOTP, benefits over passwords | PM, Stakeholders | 2KB |
| [design.md](design.md) | Database schema, architecture, data flow | Architects, Dev | 12KB |
| [IMPLEMENTATION-STATUS.md](IMPLEMENTATION-STATUS.md) | Implementation phases, status, metrics | PM, Dev Lead | 8KB |

### Implementation Guides

| Document | Purpose | Audience | Size |
|----------|---------|----------|------|
| [TOTP-MIGRATION.md](TOTP-MIGRATION.md) | Backend migration checklist, API endpoints | Backend Dev | 8KB |
| [QUICKSTART.md](QUICKSTART.md) | Code examples, common pitfalls, quick reference | All Devs | 8KB |
| [UI-MIGRATION-TANSTACK.md](UI-MIGRATION-TANSTACK.md) | Frontend with TanStack, components, hooks | Frontend Dev | 42KB |

### Specification Files

| Document | Purpose | Endpoints Covered |
|----------|---------|-------------------|
| [specs/totp-authentication/spec.md](specs/totp-authentication/spec.md) | Core TOTP generation, verification, encryption | `/totp/setup`, `/totp/verify` |
| [specs/signin/spec.md](specs/signin/spec.md) | TOTP login flow, session management | `/login`, `/login/backup-code` |
| [specs/signup/spec.md](specs/signup/spec.md) | TOTP setup during signup, backup codes | `/totp/setup`, `/totp/verify-setup` |
| [specs/totp-reset-recovery/spec.md](specs/totp-reset-recovery/spec.md) | Recovery without passwords | `/recovery/request`, `/operators/:id/totp-reset` |
| [specs/whatsapp-invitations/spec.md](specs/whatsapp-invitations/spec.md) | WhatsApp invitations with TOTP | `/invitations`, `/invitations/accept` |
| [specs/email-verification/spec.md](specs/email-verification/spec.md) | Email verification (admin-only) | `/verify-email` |

---

## 🚀 Implementation Phases Overview

### Backend (7 Phases)

**Phase 1: Foundation** (Week 1-2)
- Database migrations
- TOTP library integration
- Encryption utilities
- WhatsApp templates

**Phase 2: TOTP Signup** (Week 2-3)
- Secret generation endpoint
- QR code generation
- Backup codes generation
- Frontend TOTP setup UI

**Phase 3: TOTP Login** (Week 3-4)
- TOTP login endpoint
- Backup code login
- Session management
- Rate limiting

**Phase 4: Recovery** (Week 4-5)
- Admin TOTP reset
- Recovery request flow
- Audit logging

**Phase 5-7: Polish & Launch** (Week 5-8)
- Testing, security audit, deployment

### Frontend (6 Phases)

**Phase 1: Setup** (Week 1)
- TanStack Router + Query
- Zod schemas
- Auth context

**Phase 2: Login UI** (Week 2)
- TOTP login page
- Recovery page
- Route guards

**Phase 3: TOTP Setup UI** (Week 3)
- QR code component
- Backup codes display
- Signup integration

**Phase 4: Invitation Flow** (Week 4)
- Invitation acceptance
- Operator signup

**Phase 5: Account Management** (Week 5)
- TOTP regeneration
- Backup codes management

**Phase 6: Polish & Testing** (Week 6)
- E2E tests
- Accessibility
- Performance

---

## 🔐 Key Features Summary

### Authentication
- ✅ **TOTP-based login** - 6-digit codes from authenticator apps
- ✅ **Backup codes** - 10 single-use codes for recovery
- ✅ **No passwords** - Eliminates password breach risk
- ✅ **Session management** - 8h default, 30d remember-me option

### Onboarding
- ✅ **WhatsApp invitations** - Primary channel for operators
- ✅ **QR code setup** - Scan with Google Authenticator, Authy, etc.
- ✅ **Manual entry** - Fallback for QR code scanning
- ✅ **Email optional** - Operators can use WhatsApp only

### Recovery
- ✅ **Self-service** - Use backup codes
- ✅ **Admin reset** - Contact admin if no backup codes
- ✅ **WhatsApp notifications** - Instant recovery alerts
- ✅ **Audit logging** - Complete recovery event history

### Security
- ✅ **Encrypted secrets** - AES-256-GCM encryption at rest
- ✅ **Bcrypt-hashed codes** - Single-use backup codes
- ✅ **Constant-time comparison** - Timing attack prevention
- ✅ **Rate limiting** - Prevent brute force attacks

---

## 📈 Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| Operator onboarding time | < 2 minutes | Time from invitation click to dashboard |
| TOTP setup success rate | > 90% | Successful setups / attempts |
| WhatsApp delivery rate | > 95% | Messages delivered / sent |
| Invitation acceptance | > 70% | Accepted invitations / sent |
| Backup code usage success | > 95% | Successful logins / attempts |
| Password incidents | 0 | Eliminated with TOTP |
| Support requests (auth) | -50% | Reduction from password resets |

---

## 🎯 Next Steps

### 1. Review Documentation (Day 1)
- [ ] Read [README.md](README.md) for overview
- [ ] Review [QUICKSTART.md](QUICKSTART.md) for code examples
- [ ] Study [specs/totp-authentication/spec.md](specs/totp-authentication/spec.md) for TOTP details

### 2. Setup Infrastructure (Day 2-3)
- [ ] Choose libraries (see [QUICKSTART.md](QUICKSTART.md#recommended-go-libraries))
- [ ] Generate TOTP encryption key
- [ ] Install frontend dependencies (TanStack, etc.)

### 3. Run Phase 1 (Week 1-2)
- [ ] Backend: Database migrations
- [ ] Backend: TOTP library integration
- [ ] Frontend: TanStack Router setup

### 4. Continue with Phases 2-7
Follow detailed tasks in [TOTP-MIGRATION.md](TOTP-MIGRATION.md) or [UI-MIGRATION-TANSTACK.md](UI-MIGRATION-TANSTACK.md)

---

## ❓ FAQ

**Q: Why TOTP instead of passwords?**  
A: TOTP eliminates password breach risk, removes password reset complexity, and provides superior security with better UX.

**Q: What if users lose their phone?**  
A: Users have 10 backup codes. If those are lost too, admin can reset TOTP via WhatsApp.

**Q: Do operators need email?**  
A: No! Operators can sign up with WhatsApp number only. Email is optional for operators.

**Q: Which authenticator apps are supported?**  
A: Any TOTP-compliant app: Google Authenticator, Authy, 1Password, Bitwarden, LastPass, etc.

**Q: How long does implementation take?**  
A: 6-8 weeks with a team of 2-3 developers (see [IMPLEMENTATION-STATUS.md](IMPLEMENTATION-STATUS.md))

**Q: Can we migrate existing users?**  
A: Yes! See [TOTP-MIGRATION.md](TOTP-MIGRATION.md#migration-strategy) for phased migration plan.

---

## 🔗 Related Documentation

- [Operator Dashboard Spec](../../operator-dashboard/proposal.md)
- [WhatsApp Integration](../../../README.md)
- [API Documentation](../../../docs/api.md) (if exists)
- [Database Schema](../../../migrations/) (existing migrations)

---

## 📞 Support

For questions about this specification:
1. Check the relevant spec file in `specs/`
2. Review [QUICKSTART.md](QUICKSTART.md) for code examples
3. Consult [TOTP-MIGRATION.md](TOTP-MIGRATION.md) for implementation details

**Ready to start? Begin with Phase 1 in [TOTP-MIGRATION.md](TOTP-MIGRATION.md)!** 🚀
