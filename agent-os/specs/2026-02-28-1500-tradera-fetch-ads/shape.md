## Shaping Notes

- Scope: Guarantee Tradera buy-now detection and ensure emails are only sent when buy-now exists.
- Non-goals: UI changes, notification channels beyond email, support for new marketplaces.
- Decisions:
  - Keep `shouldSendTradingRuleEmail` policy: for Tradera, only when `HasBuyNow == true`.
  - Enrich API results with GetItem calls only when Search response lacks buy-now info; avoid calling GetItem for every item unconditionally.
  - Shared ad struct fields will be authoritative across packages to avoid mismatches.
- Constraints: Tradera SOAP API MaxAPICallsPerDay (currently 80) and 2s min delay — must be respected in enrichment.
