---
description: Inject relevant standards into current context
agent: build
---

Inject relevant standards into the current context, formatted appropriately for the situation.

## Usage Modes

### Auto-Suggest Mode (no arguments)

Analyze context and suggest relevant standards.

### Explicit Mode (with arguments)

- `inject-standards api` — All standards in api/
- `inject-standards api/response-format` — Single file
- `inject-standards api/response-format api/auth` — Multiple files
- `inject-standards root` — All standards in the root folder
- `inject-standards root/naming` — Single file from root folder

## Process

### Step 1: Detect Context Scenario

Determine which scenario we're in:

1. **Conversation** — Regular chat, implementing code, answering questions
2. **Creating a Skill** — Building a `.opencode/skills/` file
3. **Shaping/Planning** — Building a spec, planning features

If uncertain, ask user to confirm.

### Step 2: Read the Index (Auto-Suggest Mode)

Read `agent-os/standards/index.yml` to get the list of available standards and their descriptions.

### Step 3: Analyze Work Context

Look at the current conversation to understand what the user is working on.

### Step 4: Match and Suggest

Match index descriptions against the context. Present suggestions and let user select.

### Step 5: Inject Based on Scenario

**Conversation:** Read the standards and announce them with full content.

**Creating a Skill:** Ask if user wants file references or copied content.

**Shaping/Planning:** Ask if user wants file references or copied content for the plan.

## Explicit Mode

When arguments are provided, skip the suggestion step but still detect scenario. Validate that specified files/folders exist.

## Tips

- Run early — Inject standards at the start of a task, before implementation
- Be specific — If you know which standards apply, use explicit mode
- Check the index — If suggestions seem wrong, run `index-standards` to rebuild
