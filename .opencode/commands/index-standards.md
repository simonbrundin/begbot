---
description: Rebuild and maintain the standards index file
agent: build
---

Rebuild and maintain the standards index file (`index.yml`). The index enables `inject-standards` to suggest relevant standards without reading all files.

## Process

### Step 1: Scan for Standards Files

List all `.md` files in `agent-os/standards/` and its subfolders.

### Step 2: Load Existing Index

Read `agent-os/standards/index.yml` if it exists. Note which entries already have descriptions.

### Step 3: Identify Changes

Compare the file scan with the existing index:

- **New files** — Standards files without index entries
- **Deleted files** — Index entries for files that no longer exist
- **Existing files** — Already indexed, keep as-is

### Step 4: Handle New Files

For each new standard file, read it and propose a short description. Keep descriptions to one short sentence.

### Step 5: Handle Deleted Files

Remove index entries for files that no longer exist automatically. Report what was removed.

### Step 6: Write Updated Index

Generate `agent-os/standards/index.yml` with this structure:

```yaml
folder-name:
  file-name:
    description: Brief description here
```

**Rules:**

- Alphabetize folders
- Alphabetize files within each folder
- File names without `.md` extension
- One-line descriptions only

### Step 7: Report Results

Summarize what changed (new entries added, stale entries removed, unchanged entries).

## When to Run

- After manually creating or deleting standards files
- If `inject-standards` suggestions seem out of sync
- To clean up a messy or outdated index
