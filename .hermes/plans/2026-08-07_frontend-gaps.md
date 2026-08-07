# Frontend Gap Implementation Plan

> **Context:** 38 backend patient sub-resource routes exist, but only 7 have frontend pages. Plus new admin/billing modules from gap analysis need UI.

---

## Phase 1: HIGH — Clinical Patient Sub-Screens (12 pages)

All follow the same pattern: Svelte 5 + Bootstrap 5 table with list/create, same API helper, same component imports. Highly mechanical — can dispatch 3 subagents in parallel for 4 pages each.

### Batch 1A: Core Clinical (4 pages)
- `patients/[id]/immunizations/+page.svelte` — table of immunizations, add new form
- `patients/[id]/referrals/+page.svelte` — referral list, create referral form
- `patients/[id]/labs/+page.svelte` — lab results table
- `patients/[id]/problems/+page.svelte` — tabbed view: current-problems + chronic-problems, add form

### Batch 1B: Surgical + Certifications (4 pages)
- `patients/[id]/surgical-history/+page.svelte` — surgical history table + add
- `patients/[id]/certifications/+page.svelte` — certifications table + add
- `patients/[id]/financial/+page.svelte` — financial demographics table + add
- `patients/[id]/episodes-of-care/+page.svelte` — EOC list + create

### Batch 1C: More Clinical (4 pages)
- `patients/[id]/drug-samples/+page.svelte` — drug sample inventory
- `patients/[id]/clinical-orders/+page.svelte` — clinical orders list + create
- `patients/[id]/authorizations/+page.svelte` — auth list + create
- `patients/[id]/growth-charts/+page.svelte` — growth chart data table

---

## Phase 2: MEDIUM — Admin + Billing Pages (6 pages)

### Batch 2A: Admin (3 pages)
- `admin/user-groups/+page.svelte` — manage user→group assignments (reuse ACL admin layout)
- `admin/user-preferences/+page.svelte` — preference key/value editor
- `admin/form-templates/+page.svelte` — JSON template editor

### Batch 2B: Billing + Clinical (3 pages)
- `billing/claimlog/+page.svelte` — claims audit log viewer
- `patients/[id]/letters/+page.svelte` — letter list + create
- `patients/[id]/correspondence/+page.svelte` — correspondence list + compose

---

## Phase 3: LOW — Remaining Sub-Screens (11 pages)

### Batch 3A: Quick Wins (4 pages)
- `patients/[id]/coverage-info/+page.svelte` — coverage list
- `patients/[id]/procedures/+page.svelte` — procedure list
- `patients/[id]/ledger/+page.svelte` — financial ledger with date range
- `patients/[id]/payments/+page.svelte` — payment history

### Batch 3B: Remaining (4 pages — read-only tables)
- `patients/[id]/phones/+page.svelte`
- `patients/[id]/photo-id/+page.svelte`
- `patients/[id]/signatures/+page.svelte`
- `patients/[id]/addresses/+page.svelte`

### Batch 3C: Final (3 pages — read-only)
- `patients/[id]/tags/+page.svelte`
- `patients/[id]/annotations/+page.svelte`
- `patients/[id]/claims/+page.svelte`

---

## Phase 4: Navigation Updates

- Update patient sidebar/nav in `patients/[id]/+page.svelte` to include links to all new sub-pages
- Group by category: Clinical, Financial, Administrative
- Add admin nav entries for user-groups, user-preferences, form-templates
- Add billing nav entries for claimlog

---

## Execution

```
Phase 1:  1A (4) → 1B (4) → 1C (4)    [3 subagent batches, 4 pages each]
Phase 2:  2A (3) + 2B (3) parallel     [2 subagents, 3 pages each]
Phase 3:  3A (4) → 3B (4) → 3C (3)    [3 batches]
Phase 4:  Single subagent              [navigation wiring]
```

**Total: ~29 new SvelteKit pages across 8 dispatches.** All pages follow identical patterns — reusable `LoadingSpinner`, `ErrorBanner`, `EmptyState` components already exist. Estimated 15-20 minutes per batch.
