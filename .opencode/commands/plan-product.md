---
description: Establish foundational product documentation
agent: build
---

Establish foundational product documentation through an interactive conversation. Creates mission, roadmap, and tech stack files in `agent-os/product/`.

## Process

### Step 1: Check for Existing Product Docs

Trigger notification:
```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_input "Checking existing product documentation"
```

Check if `agent-os/product/` exists. If any files exist, ask if user wants to:

1. Start fresh (replace all)
2. Update specific files
3. Cancel

### Step 2: Gather Product Vision (for mission.md)

Trigger notification:
```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_input "Gathering product vision"
```

Ask:

- **What problem does this product solve?**
- **Who is this product for?**
- **What makes your solution unique?**

### Step 3: Gather Roadmap (for roadmap.md)

Trigger notification:
```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_input "Gathering roadmap information"
```

Ask:

- **What are the must-have features for launch (MVP)?**
- **What features are planned for after launch?**

### Step 4: Establish Tech Stack (for tech-stack.md)

Trigger notification:
```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_input "Establishing tech stack"
```

Check if `agent-os/standards/global/tech-stack.md` exists. If it does, ask if this project uses the same tech stack or differs.

If different or no standard exists, ask user to specify:

- Frontend
- Backend
- Database
- Other (hosting, APIs, tools, etc.)

### Step 5: Generate Files

Create the `agent-os/product/` directory and generate each file:

**mission.md:**

```markdown
# Product Mission

## Problem

[Insert what problem this product solves]

## Target Users

[Insert who this product is for]

## Solution

[Insert what makes the solution unique]
```

**roadmap.md:**

```markdown
# Product Roadmap

## Phase 1: MVP

[Insert must-have features for launch]

## Phase 2: Post-Launch

[Insert planned future features]
```

**tech-stack.md:**

```markdown
# Tech Stack

## Frontend

[Frontend technologies, or "N/A"]

## Backend

[Backend technologies, or "N/A"]

## Database

[Database choice, or "N/A"]

## Other

[Other tools, hosting, services]
```

Trigger notification when complete:
```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_complete "Product documentation created"
```
