---
description: Extract tribal knowledge from codebase into documented standards
agent: build
---

Extract tribal knowledge from your codebase into concise, documented standards.

## Process

### Step 1: Determine Focus Area

Analyze the codebase structure (folders, file types, patterns). Identify 3-5
major areas. Examples:

- **Frontend areas:** UI components, styling/CSS, state management, forms,
  routing
- **Backend areas:** API routes, database/models, authentication, background
  jobs
- **Cross-cutting:** Error handling, validation, testing, naming conventions,
  file structure

Use question tool to present the areas and let user choose.

### Step 2: Analyze & Present Findings

Read key files in the selected area (5-10 representative files). Look for
patterns that are:

- **Unusual or unconventional** — Not standard framework/library patterns
- **Opinionated** — Specific choices that could have gone differently
- **Tribal** — Things a new developer wouldn't know without being told
- **Consistent** — Patterns repeated across multiple files

Present findings and let user select which to document.

### Step 3: Ask Why, Then Draft Each Standard

For each selected standard, ask 1-2 clarifying questions about the "why" behind
the pattern. Draft the standard incorporating their answer. Confirm with user
before creating the file.

Example questions:

- "What problem does this pattern solve? Why not use the default/common
  approach?"
- "Are there exceptions where this pattern shouldn't be used?"
- "What's the most common mistake a developer or agent makes with this?"

Process one standard at a time through the full loop.

### Step 4: Create the Standard File

For each standard (after completing Step 3's Q&A):

1. Determine the appropriate folder (create if needed): `api/`, `database/`,
   `javascript/`, `css/`, `backend/`, `testing/`, `global/`
2. Check if a related standard file already exists — append to it if so
3. Draft the content and confirm with user
4. Create or update the file in `agent-os/standards/[folder]/`

### Step 5: Update the Index

After all standards are created, scan `agent-os/standards/` for all `.md` files.
For each new file without an index entry, propose a description and update
`agent-os/standards/index.yml`:

```yaml
api:
  response-format:
    description: API response envelope structure and error format
```

Alphabetize by folder, then by filename.

### Step 6: Offer to Continue

Ask if user wants to discover standards in another area.

## Writing Concise Standards

- **Lead with the rule** — State what to do first, explain why second (if
  needed)
- **Use code examples** — Show, don't tell
- **Skip the obvious** — Don't document what the code already makes clear
- **One standard per concept** — Don't combine unrelated patterns
- **Bullet points over paragraphs** — Scannable beats readable
