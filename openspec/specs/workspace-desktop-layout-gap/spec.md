# workspace-desktop-layout-gap Specification

## Purpose

Eliminate the large blank gap above the "WhatsApp Workspace" header on desktop
viewports. On screens at or above the `lg` breakpoint the sticky header and all
page content render roughly `631px` down the page because the left sidebar is
placed in normal document flow instead of being pinned to the viewport.

Root cause: in `frontend/src/components/Layout.tsx` the sidebar wrapper declares
`fixed inset-y-0 left-0 w-64 ... lg:static`. At `>= lg` the `lg:static` class
wins, switching the sidebar from `position: fixed` to `position: static`. The
sidebar (whose natural content height is ~`631px`) then becomes a regular
block-level element and the sibling content area — `lg:pl-64 min-h-screen flex
flex-col` — stacks *below* it, pushing the sticky "WhatsApp Workspace" header
and every page down by the sidebar's full height. On mobile/tablet the layout
was unaffected because `lg:static` never applies.

## Requirements

### Requirement: The sidebar stays pinned to the left edge of the viewport on desktop

At and above the `lg` breakpoint the sidebar MUST remain `position: fixed`,
spanning the full viewport height (`inset-y-0`) on the left edge, with the
content column offset to the right via `lg:pl-64`.

#### Scenario: Desktop viewport renders the workspace header at the top

- **WHEN** the authenticated `Layout` renders with a viewport of `1440x900`
- **THEN** the sidebar wrapper MUST have a computed `position` of `fixed`
  spanning the full viewport height
- **AND** the top of the sticky "WhatsApp Workspace" header MUST line up with
  the top of the content column (`offsetTop`/bounds `top` of `0`)
- **AND** no empty vertical gap MUST appear between the top of the viewport and
  the header

### Requirement: The desktop sidebar uses fixed positioning, never static

The class that turns the sidebar into a normal-flow block element at `lg`
MUST be removed. `lg:static` in the sidebar wrapper MUST be replaced with a
behavior that keeps the sidebar fixed (`lg:fixed`) on desktop while preserving
the mobile drawer, which is revealed by toggling `translate-x-0`.

#### Scenario: Sidebar wrapper class list asserts no static override

- **WHEN** the sidebar wrapper element is rendered with the lg variant applied
- **THEN** its classes MUST NOT include a `static` positioning override
- **AND** its classes MUST include a `fixed` positioning behavior at `lg`

### Requirement: Mobile and tablet layouts keep the existing drawer behavior

Below the `lg` breakpoint the sidebar MUST remain hidden off-screen
(`-translate-x-full`) by default and slide in via `translate-x-0` when the menu
button is toggled, unchanged by this fix.

#### Scenario: Mobile drawer still hides and reveals

- **WHEN** a viewport below `lg` loads the authenticated `Layout` with the
  sidebar closed
- **THEN** the sidebar MUST be translated fully off the left edge of the screen
- **AND** opening the menu MUST translate the sidebar into view

### Requirement: The fix is covered by a regression test

A regression test MUST guard against the sidebar being pushed into normal flow.
Because the test environment (`jsdom`) does not compute real layout, the test
MUST assert the rendered sidebar wrapper classes (no `static` positioning at
`lg`) or use an equivalent marker that fails when the bug is reintroduced.

#### Scenario: Regression test fails when `lg:static` is reintroduced

- **WHEN** a test renders `Layout` with an authenticated user and inspects the
  sidebar wrapper
- **THEN** the assertion MUST fail if `lg:static` (or any static positioning
  override) is present in the sidebar classes
- **AND** the full frontend test suite (`npm test`) and type check
  (`npx tsc --noEmit`) MUST pass

## Notes

- The verified fix changes `lg:static` to `lg:fixed` in
  `frontend/src/components/Layout.tsx:57`.
- Measured before the fix (Chromium, `1440x900`): header `top: 631`,
  `documentHeight: 1531`. After the fix: header `top: 0`, sidebar `position:
  fixed` spanning `0-900`, `documentHeight: 900`. Tablet (`900px`) and mobile
  (`390px`) were already correct and remain unchanged.