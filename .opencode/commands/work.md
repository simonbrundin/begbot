---
description: Check for specs to work on and select one
agent: build
---

Check if there are any specs to work on and help select one.

## Step 1: Find Existing Specs

List all folders in `agent-os/specs/`. Each folder represents a spec.

## Step 2: Check for Pending Work

For each spec folder, check if there's a `plan.md` file and look for incomplete tasks. A spec is considered "ready to work on" if:
- It has a `plan.md` with tasks
- At least one task is not marked as complete

## Step 3: Present Options to User

If there are specs with pending work:

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

## Step 4: Load Selected Spec

Once the user selects a spec:

1. Read `agent-os/specs/{selected-spec}/plan.md`
2. Read `agent-os/specs/{selected-spec}/shape.md` for context
3. Present the first incomplete task and ask for confirmation to start

## Step 5: Begin Work

After user confirms, set the current working context to the selected spec and begin the first pending task.
