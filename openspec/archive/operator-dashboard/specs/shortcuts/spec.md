# Keyboard shortcuts requirements for operator dashboard

## ADDED Requirements

### Requirement: Common actions have keyboard shortcuts

Frequent operator actions MUST be accessible via keyboard.

#### Scenario: Navigate inbox

- **WHEN** focus is in the inbox
- **THEN** arrow keys MUST navigate between conversations
- **AND** Enter MUST open the selected conversation

#### Scenario: Send reply

- **WHEN** focus is in the reply composer
- **THEN** Ctrl+Enter or Cmd+Enter MUST send the message
- **AND** Shift+Enter MUST insert a newline

### Requirement: Shortcuts are discoverable

A shortcut reference MUST be available to users.

#### Scenario: Shortcut help modal

- **WHEN** a user presses `?` or clicks Help
- **THEN** a modal MUST display all available shortcuts
- **AND** group them by context (inbox, conversation, global)

### Requirement: Shortcuts do not conflict with assistive tech

All shortcuts MUST avoid conflicts with screen reader and OS commands.

#### Scenario: Screen reader compatibility

- **WHEN** a screen reader is active
- **THEN** custom shortcuts MUST not interfere with screen reader navigation
- **AND** SHOULD use modifier keys (Ctrl, Alt, Cmd) to avoid conflicts
