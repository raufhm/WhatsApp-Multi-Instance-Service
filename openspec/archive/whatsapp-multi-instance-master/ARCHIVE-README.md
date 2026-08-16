# Archived: WhatsApp Multi-Instance Service — Master Specification

**Archived**: August 16, 2026
**Status**: 📚 REFERENCE DOCUMENT (not a feature change)

This is the complete architectural specification for the service. It is a
reference document, not an implementable change, and is preserved here for
historical/onboarding context.

## What This Is

`master-specification.md` documents the system architecture:

- Multi-instance WhatsApp management
- Bot/rule engine
- Operator dashboard
- Conversation management (lifecycle, merge/split)
- Media & upload management (S3 + local disk)
- TOTP authentication
- Data model, API surface, configuration, and deployment

## Note

This was kept under `openspec/changes/` as a living architecture doc. It is
archived here since it is not an active change. The concrete, implementable
changes it describes are archived separately:

- `../operator-dashboard/`
- `../operator-permissions-completed/`
- `../tenant-onboarding-flow-completed/`
- `../media-storage-and-delivery-completed/`
- `../reliable-s3-media-uploads-completed/`
- `../manual-onboarding-whatsapp-completed/`

## Related

- `../agent-implementation-plan/` — prioritized backlog of remaining gaps
- `../../../AGENT.md` — original product plan
