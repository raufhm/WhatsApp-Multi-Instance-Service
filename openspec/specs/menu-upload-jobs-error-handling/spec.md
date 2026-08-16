# menu-upload-jobs-error-handling Specification

## Purpose

Fix the left-nav **Upload Jobs** menu so operators never see an unhandled
failure ("Something went wrong!") when opening the page. Today `UploadJobs.tsx`
(`frontend/src/pages/UploadJobs.tsx:29`) loads via `useUploadJobs(status)` →
`GET /dashboard/api/upload-jobs`; on failure it shows only
"Failed to load upload jobs." with no Retry control, and the selected status
filter is kept while the error text replaces the whole table.

## Requirements

### Requirement: The Upload Jobs failure state shows an actionable error with Retry

When the upload-jobs query fails, the page MUST render a specific message with
a Retry button that calls `refetch()`, while preserving the currently selected
status filter.

#### Scenario: Upload jobs request fails

- **WHEN** `GET /dashboard/api/upload-jobs` returns a non-2xx response or a
  network error
- **THEN** the page MUST NOT show a generic "Something went wrong!" message
- **AND** it MUST show a specific error message with a Retry button that
  triggers `refetch()`
- **AND** the status filter dropdown MUST keep its selected value so a retry
  re-requests the same filter

### Requirement: Loading, empty, and data states remain intact

The loading spinner, the "No upload jobs found" empty state, and the status
badge table MUST continue to render correctly per the selected filter.

#### Scenario: Page renders with zero upload jobs

- **WHEN** `GET /dashboard/api/upload-jobs` succeeds with an empty array
- **THEN** the page MUST render the "No upload jobs found" empty state

### Requirement: The fix is covered by regression tests

#### Scenario: Regression test guards the retry path

- **WHEN** a frontend test renders `UploadJobs` against a failing query
- **THEN** the test MUST find a Retry button that refetches with the same
  status filter

## Notes

- `useUploadJobs` polls every 15s (`frontend/src/hooks/useDashboard.ts:67`);
  the error state must coexist with polling without flicker.
- After source changes, rebuild the embedded frontend (`npm run build`) so the
  served bundle no longer contains the generic "Something went wrong!" surface.