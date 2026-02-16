# Shape: Nypris Display Format

## Scope
Small UI adjustment to improve information hierarchy on the ads listing page.

## Context
The `/ads` page displays listings with various price information. Currently:
- Nypris (original retail price) is shown above the listing price
- Frakt (shipping cost) and Värdering (valuation) are shown below the listing price

This creates a disjointed reading experience where related price information is separated.

## Decision
Move Nypris to be grouped with Frakt and Värdering, creating a consistent "price details" section that appears after the main listing price.

## Constraints
- Must maintain existing Swedish text
- Must use existing styling patterns
- Must handle null/undefined values gracefully
- No backend changes required (data already available)

## Out of Scope
- Changing Delvärderingar display (intentionally in boxes)
- Changing any other UI elements
- Adding new data fields
