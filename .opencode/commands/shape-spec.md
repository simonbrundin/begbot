---
description: Gather context and structure planning for significant work
agent: build
---

Gather context and structure planning for significant work. **Run this command
when planning features.**

## Process

### Step 1: Clarify What We're Building

Trigger notification before asking clarifying questions:

```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_input "Clarifying feature scope"
```

Understand the scope. Ask 1-2 clarifying questions if the scope is unclear:

- "Is this a new feature or a change to existing functionality?"
- "What's the expected outcome when this is done?"
- "Are there any constraints or requirements I should know about?"

### Step 2: Gather Visuals

Trigger notification before asking about visuals:

```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_input "Gathering visual references"
```

Ask if user has any visuals to reference:

- Mockups or wireframes
- Screenshots of similar features
- Examples from other apps

If visuals are provided, note them for inclusion in the spec folder.

### Step 3: Identify Reference Implementations

Trigger notification before asking about references:

```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_input "Identifying reference implementations"
```

Ask if there's similar code in this codebase to reference. If references are
provided, read and analyze them.

### Step 4: Check Product Context

Check if `agent-os/product/` exists and contains files. If it exists, read key
files and ask if this feature should align with any specific product goals or
constraints.

### Step 5: Surface Relevant Standards

Read `agent-os/standards/index.yml` to identify relevant standards based on the
feature being built. Confirm which to include.

### Step 6: Generate Spec Folder Name

Create a folder name using this format:

```
YYYY-MM-DD-HHMM-{feature-slug}/
```

Example: `2026-01-15-1430-user-comment-system/`

### Step 7: Structure the Plan

Build the plan with **Task 1 always being "Save spec documentation"**.

Create `agent-os/specs/{folder-name}/` with:

- **plan.md** — This full plan
- **shape.md** — Shaping notes (scope, decisions, context)
- **standards.md** — Relevant standards that apply
- **references.md** — Pointers to reference implementations
- **visuals/** — Any mockups or screenshots

### Step 8: Complete the Plan

Build out the remaining implementation tasks based on:

- The feature scope
- Patterns from reference implementations
- Constraints from standards

Trigger notification when spec planning is complete:

```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_complete "Spec planning complete: $spec_name"
```

## Output Structure

```
agent-os/specs/{YYYY-MM-DD-HHMM-feature-slug}/
├── plan.md           # The full plan
├── shape.md          # Shaping decisions and context
├── standards.md      # Which standards apply
├── references.md     # Pointers to similar code
└── visuals/          # Mockups, screenshots
```
