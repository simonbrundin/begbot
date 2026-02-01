# Work on Spec

Check for specs with pending work and select one to continue.

## Important Guidelines

- **Always use AskUserQuestion tool** when asking the user anything
- **Use the question tool** for numbered selection
- **Be helpful** — summarize what each spec is about when presenting options

## Step 1: Find Existing Specs

List all folders in `agent-os/specs/`. Each folder represents a spec with format: `YYYY-MM-DD-HHMM-feature-slug/`

## Step 2: Check for Pending Work

For each spec folder, check if there's a `plan.md` file. Look for incomplete tasks — tasks not marked as complete/done.

A spec is "ready to work on" if it has a `plan.md` with at least one incomplete task.

## Step 3: Present Options to User

If there are specs with pending work:

Use the **question** tool to ask the user which spec to work on:

```
Which spec would you like to work on?

1. 2026-01-30-1430-openrouter-model-per-function
2. 2026-01-30-0900-listing-product-validation
3. 2026-01-30-0921-search-terms-table

(Enter the number of your choice)
```

If there are NO specs with pending work:

Tell the user:
```
No specs with pending work found.

Consider running /shape-spec to create a new spec and plan new work.
```

## Step 4: Load Selected Spec

Once the user selects a spec:

1. Read `agent-os/specs/{selected-spec}/plan.md`
2. Read `agent-os/specs/{selected-spec}/shape.md` for context
3. Identify the first incomplete task
4. Present it to the user and ask for confirmation to start

Example:
```
Selected: 2026-01-30-1430-openrouter-model-per-function

First pending task:
**Task 2: Implement model selection logic**

[Task description from plan.md]

Ready to start? (confirm / skip / another task)
```

## Step 5: Begin Work

After user confirms, proceed with the selected task and update progress in the plan.
