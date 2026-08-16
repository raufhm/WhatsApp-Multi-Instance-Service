# crm-under-contacts-activity Specification

## Purpose

Add a CRM experience to the dashboard that lives **under the existing Contacts
left-nav menu** (it is not a separate top-level menu). Operators land on the
contact directory at `/contacts` and open the CRM record for any contact at
`/contacts/$id`. Each contact record MUST have **its own activity feed** so an
operator sees follow-ups and customer actions per person, not only the
conversation-wide queue shown on the Inbox.

Today `activity` rows are scoped to conversations
(`backend/domain/models.go:196-204`, `Activity.ConversationID`), there is no
contact-detail route, and the contact payload is not serialized in the schema
the frontend expects (`menu-contacts-error-handling` covers that gap). This
spec makes Contacts function as a CRM: directory → contact profile → per-contact
activity timeline.

## Requirements

### Requirement: The CRM surface is navigated from the Contacts menu

The left-nav **Contacts** item MUST remain the single entry point. Selecting a
contact in the directory opens `/contacts/$id`, a CRM record showing contact
profile data, a per-contact activity timeline, and links to the contact's
conversations.

#### Scenario: Operator opens a contact's CRM record

- **WHEN** an operator opens `/contacts` and selects a contact card
- **THEN** the browser navigates to `/contacts/$id` with that contact's id
- **AND** the CRM record MUST render the contact name, number, email, and tags

### Requirement: Each contact has its own activity feed

Activities MUST be addressable per contact. The `activities` table gains a
nullable `contact_id` column; when an activity is created from a conversation,
`contact_id` is derived from that conversation's `contact_id`. A new endpoint
`GET /api/v1/contacts/{id}/activities` MUST return the contact's activities
ordered newest-first, including contact-level activities created directly.

#### Scenario: Contact activity timeline is loaded

- **WHEN** an operator opens `/contacts/$id`
- **THEN** the CRM record MUST fetch `GET /api/v1/contacts/{id}/activities`
- **AND** display each activity's type, summary, status/priority, and due date
  newest-first
- **AND** activities created under any of the contact's conversations MUST
  appear in the same timeline

### Requirement: Operators can create contact-level follow-up activities

A CRM record MUST support creating a contact-level follow-up via
`POST /api/v1/contacts/{id}/activities` with `type`, `summary`, `next_action`,
`priority`, and optional `due_at`; the new activity is stored with that
contact's id.

#### Scenario: Operator schedules a follow-up for a contact

- **WHEN** an operator posts a follow-up on a contact's CRM record
- **THEN** the activity is created with `contact_id` set to that contact
- **AND** the timeline refreshes to include the new pending activity

### Requirement: Contact serialization supports the CRM surface

`GET /api/v1/contacts` and `GET /api/v1/contacts/{id}` MUST return the
display-facing fields used by the directory and CRM record (`id`, `tenant_id`,
`name`, `number`, `email`, `tags`, `created_at`, `updated_at`), as specified by
`menu-contacts-error-handling`.

#### Scenario: Contact record returns schema-compatible JSON

- **WHEN** `GET /api/v1/contacts/{id}` succeeds for a known contact
- **THEN** the payload MUST contain `name`, `number`, `email`, and `tags`
- **AND** the CRM record MUST render each field from the response

### Requirement: CRM pages handle loading, empty, and error states

The directory (`/contacts`) and CRM record (`/contacts/$id`) MUST show loading
spinners, a clear empty state when a contact has no activities, and an
actionable error with Retry when their queries fail (consistent with the
`menu-contacts-error-handling` retry handling).

#### Scenario: Contact has no activities yet

- **WHEN** `GET /api/v1/contacts/{id}/activities` succeeds with an empty list
- **THEN** the CRM record MUST render an empty-timeline message inviting a
  follow-up
- **AND** the follow-up composer MUST remain available

### Requirement: The feature is covered by backend and frontend tests

#### Scenario: Regression tests cover contact-scoped activities

- **WHEN** a backend test creates activities for a contact across two
  conversations and calls `GET /api/v1/contacts/{id}/activities`
- **THEN** the endpoint MUST return both activities with the contact id set
- **AND** a frontend test for `/contacts/$id` MUST find the activity timeline
  and the follow-up composer

## Notes

- Navigation model: keep a single left-nav "Contacts" item; do not add a
  separate "CRM" menu yet. Rename the item label to "CRM" later only if the
  module expands beyond contacts/activity.
- `contacts.metadata` (`backend/domain/models.go:54`) already stores arbitrary
  fields; `name`/`email`/`tags` should be read from `DisplayName` +
  `metadata`.
- Migration required: `ALTER TABLE activities ADD COLUMN contact_id UUID NULL`
  plus a backfill from `conversations.contact_id`.
- Existing Inbox activity queue (`frontend/src/pages/Inbox.tsx:204-243`)
  remains the conversation-scoped view; the CRM timeline is the
  contact-scoped view.