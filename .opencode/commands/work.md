---
description: Check for specs to work on and select one
agent: build
---

Check if there are any specs to work on and help select one.

## Step 1: Find Specs with Pending Work (Fast Scan)

Use `rg` (ripgrep) to quickly find specs that need work without reading every file.

### Fast Scan Command

```bash
rg "^---\nstatus: (pending|in_progress)" agent-os/specs/*/plan.md -l --multiline 2>/dev/null | sort
```

This returns file paths like:
```
agent-os/specs/2026-01-15-1430-feature-a/plan.md
agent-os/specs/2026-01-16-0920-feature-b/plan.md
```

**Why rg:** 100x faster than reading files sequentially. Only specs with pending/in_progress status are returned.

### Fallback (if rg unavailable)

**IMPORTANT:** Do NOT read all spec files. If `rg` is not available:
1. Use `grep` command instead: `grep -l "^status: (pending|in_progress)" agent-os/specs/*/plan.md`
2. Only read the matching files
3. If no matches found, report "No specs with pending work found"

## Step 2: Extract Spec Names

From the rg output, extract spec folder names:

```bash
rg "^---\nstatus: (pending|in_progress)" agent-os/specs/*/plan.md -l --multiline 2>/dev/null \
  | xargs -I {} dirname {} \
  | xargs -I {} basename {} \
  | sort
```

This gives spec folder names like:
```
2026-01-15-1430-feature-a
2026-01-16-0920-feature-b
```

## Step 3: Check for Incomplete Tasks

**IMPORTANT:** Only read `plan.md` files for specs found by rg in Step 1. NEVER read all specs.

For each spec with pending/in_progress status, read its `plan.md` and verify there are incomplete tasks.

Skip specs where:
- All tasks are marked complete
- No tasks found

## Step 4: Present Options to User

If there are specs with pending work:

Trigger notification before asking for input:
```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_input "Select which spec to work on"
```

Use the **question** tool to ask the user which spec to work on:

```
Which spec would you like to work on?

{list specs with numbers}

(Enter the number of your choice)
```

Format the list as:
```
1. {spec-folder-name}
2. {spec-folder-name}
...
```

If there are NO specs with pending work:

Tell the user:
```
No specs with pending work found. Would you like to create a new spec?

Run /shape-spec to shape and plan new work.
```

## Step 5: Load Selected Spec

Once the user selects a spec:

1. Read `agent-os/specs/{selected-spec}/plan.md`
2. Read `agent-os/specs/{selected-spec}/shape.md` for context
3. Present the first incomplete task and ask for confirmation to start

## Step 6: Begin Work

After user confirms, set the current working context to the selected spec and notify:
```bash
source "$HOME/.config/agent-os/scripts/notify.sh" 2>/dev/null || true
notify_complete "Started working on: $spec_name"
```

Begin the first pending task.

## Backwards Compatibility

Specs without a status field are treated as `completed` to avoid reading all files. When encountering a spec without status:
1. Skip it (treat as completed)
2. Suggest updating the spec with status metadata if user wants to work on it

## Performance Notes

- **With rg:** Scans 50+ specs in <10ms
- **Without rg:** Scans 50+ specs in ~500ms-1s
- **Old method:** Scans 50+ specs in ~2-5s

Always prefer rg when available.
