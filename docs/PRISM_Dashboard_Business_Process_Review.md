# PRISM Dashboard - Business Process Review

> Review date: 2026-05-06
> Scope: `/dashboard` route, current frontend implementation, and related PRISM business rules.

---

## 1. Current Dashboard Scope

`/dashboard` currently presents one planning funnel page:

- Route: `prism-frontend/src/router/routes/dashboard.routes.ts`
- Page: `prism-frontend/src/pages/dashboard/DashboardPage.vue`
- Components:
  - `prism-frontend/src/components/dashboard/DashboardKpiGrid.vue`
  - `prism-frontend/src/components/dashboard/DashboardKpiCard.vue`
  - `prism-frontend/src/components/dashboard/PlanningFunnelFlow.vue`
- Types: `prism-frontend/src/types/dashboard-flow.types.ts`

The visible business story is:

```text
Blue Book -> Green Book -> Daftar Kegiatan -> Loan Agreement
```

This matches the core PRISM planning process before loan execution monitoring. The current page is useful as an executive mockup, but it is not yet a reliable operational dashboard because the metrics and panel data are hard-coded in `DashboardPage.vue`.

### 1.1 Product Decisions After Review

The following decisions are accepted for the current implementation phase:

- `/dashboard` may keep sample data while the page is still a prototype or executive mockup. The UI must keep the sample-data label clear.
- Loan Agreement may point toward Monitoring as downstream business context, even if Monitoring is not yet modeled as a full dashboard stage.
- Dashboard access may remain gated by `bb_project.read` while a dedicated dashboard permission policy has not been introduced.

These decisions are acceptable for an interim release. They must be revisited before treating `/dashboard` as an operational reporting surface.

---

## 2. Business Process Documentation

### 2.1 Blue Book

**Business meaning**

Blue Book is the proposal and planning-entry stage. A Blue Book Project is a snapshot inside one Blue Book or revision, while the logical project identity links snapshots across revisions.

**Important rules**

- `bb_code` is unique only inside the same Blue Book.
- A project can appear again in another Blue Book revision when it represents the same logical project.
- Lender readiness starts from Lender Indication and LoI.
- LoI is a stronger signal than lender indication, but it is still not a Green Book funding source.

**Dashboard questions this stage should answer**

- How many active Blue Book Projects are in the latest or selected period?
- How many already moved to Green Book?
- How many have LoI but not Green Book?
- How many only have lender indication?
- How many have no lender indication?
- Which institutions, regions, program titles, and lenders drive the largest pipeline?

**Current dashboard coverage**

Covered partially:

- Blue Book total count and value.
- Conversion label to Green Book.
- Lender readiness categories.
- Distribution by institution type, top K/L, region, and program.

Gap:

- It is not wired to actual Blue Book, LoI, lender indication, or revision-aware latest data.

### 2.2 Green Book

**Business meaning**

Green Book is the priority funding plan stage. A Green Book Project references at least one Blue Book Project and carries funding sources, activities, disbursement plan, and funding allocation.

**Important rules**

- GB Project must reference at least one BB Project.
- Multiple BB Projects can be merged into one GB Project only when all BB Projects come from the same Blue Book header.
- One BB Project can split into more than one GB Project.
- GB Project uses latest BB Project snapshot when created or revised.
- Funding allocation follows GB activities and should stay synchronized.
- Disbursement plan is total project per year, not per lender.

**Dashboard questions this stage should answer**

- How many Green Book Projects exist for the selected publish year or latest snapshot?
- How many came from one-to-one, one-to-many, and many-to-one BB relationships?
- Which Green Book Projects are missing required readiness data?
- Which projects are not yet included in Daftar Kegiatan?
- What is the lender and currency exposure by stage?

**Current dashboard coverage**

Covered partially:

- Green Book count and value.
- Conversion label to Daftar Kegiatan.
- Lender mix, top lender, institution, and region distribution.

Gap:

- The page no longer exposes BB -> GB relationship patterns or readiness completeness, which are important business controls for this stage.

### 2.3 Daftar Kegiatan

**Business meaning**

Daftar Kegiatan is the formal activity-list stage. It freezes concrete downstream references to selected GB/BB snapshots. This is where financing detail and loan allocation become more concrete before Loan Agreement.

**Important rules**

- Final DK cannot be updated except by ADMIN.
- DK Project must use latest GB Project snapshot when created.
- Downstream references stay on the concrete snapshot used when DK was created.
- DK lender must come from BB lender indication or GB funding source.
- DK activity details are free text, not technically linked to GB activities.

**Dashboard questions this stage should answer**

- How many DK Projects have not yet become Loan Agreement?
- How old is each DK Project queue item?
- Which lenders and institutions dominate the pending Loan Agreement queue?
- Which DK records are final and blocked from normal update?
- Which DK Projects have financing detail gaps or lender mismatches?

**Current dashboard coverage**

Covered partially:

- DK count and value.
- Conversion label to Loan Agreement.
- Lender, top lender, institution, program, and region distribution.

Gap:

- The page does not show DK bottleneck age, finalization status, candidate LA queue, or financing-detail quality.

### 2.4 Loan Agreement

**Business meaning**

Loan Agreement is the legal-binding stage. One DK Project can have multiple Loan Agreements. This stage should separate signed loans, effective loans, extended loans, and loans at schedule risk.

**Important rules**

- One DK Project can have more than one Loan Agreement.
- `loan_code` is globally unique.
- `original_closing_date` is optional, but `closing_date` must be greater than or equal to it when present.
- `is_extended` and `extension_days` are computed, not stored.
- `cumulative_disbursement` is manual and follows the selected Loan Agreement currency.

**Dashboard questions this stage should answer**

- How many Loan Agreements exist and what is their total value?
- Which DK Projects still have no Loan Agreement?
- Which Loan Agreements are signed but not effective?
- Which Loan Agreements are extended?
- Which Loan Agreements are nearing closing date?

**Current dashboard coverage**

Covered partially:

- LA count and value.
- Placeholder schedule status (`On Schedule`, `Behind`, `At Risk`).
- Candidate distribution copied from DK.

Gap:

- LA values are all zero and hard-coded. It is not yet backed by `loan_agreement` data, effective/closing rules, or extension computation.

### 2.5 Monitoring Disbursement

Monitoring Disbursement is downstream execution monitoring after Loan Agreement is effective. It is part of the broader PRISM lifecycle, but it should not be mixed into the planning dashboard unless the dashboard intentionally includes an execution-monitoring tab or stage.

Current `/dashboard` title and funnel show `BB -> GB -> DK -> LA`, but the LA stage still points to `Monitoring` as the next step. This needs a product decision:

- Keep `/dashboard` as planning/legal pipeline only, ending at Loan Agreement.
- Or expand dashboard scope to include Monitoring with real monitoring metrics, explicit stage type, and supporting API contract.

---

## 3. Review Findings

### Finding 1 - Hard-coded data conflicts with real-time copy

**Status: accepted temporary constraint.**

The dashboard page defines KPI and funnel data as constants in `DashboardPage.vue`, while the page subtitle says it is a real-time PRISM snapshot. The UI also shows `Data contoh`, so the user receives two conflicting signals. This must be resolved before the dashboard is treated as operational.

**Impact**

- Users can make decisions from sample numbers.
- Counts can diverge from actual Blue Book, Green Book, DK, and LA records.
- Review of business process becomes difficult because the data lineage is not traceable.

**Improve**

- Create `dashboard.service.ts`, `dashboard.store.ts`, and `dashboard.types.ts`.
- Add backend query source under `prism-backend/sql/queries/` if dashboard data is still backend-owned.
- Expose clear endpoint(s), for example `GET /api/v1/dashboard/planning-funnel`.
- Replace sample constants with API data and loading/error/empty states.
- If the page is still a mockup, rename copy to `Data contoh` consistently and remove `real-time`.

### Finding 2 - Funnel scope is inconsistent at Loan Agreement

**Status: accepted business direction for now.**

The visible funnel is `BB -> GB -> DK -> LA`, but the LA stage has `nextLabel: 'Monitoring'`. This creates a scope mismatch because Monitoring is not represented as a real stage in the type model or page data.

**Impact**

- Users may expect Monitoring Disbursement analytics inside this dashboard.
- Product scope becomes unclear: planning/legal dashboard vs execution dashboard.
- Future implementation can accidentally mix pipeline and disbursement semantics.

**Improve**

- If dashboard remains planning/legal only, make Loan Agreement the final stage and remove Monitoring from the funnel copy.
- If Monitoring returns to dashboard scope, add it explicitly as a fifth stage with data contract, permission model, and real monitoring metrics.

### Finding 3 - Process controls are weaker than distribution charts

The current dashboard emphasizes distribution panels: lender mix, top lender, K/L, program, and region. Those are useful, but the business process needs more transition controls:

- Blue Book readiness: no lender indication, indication only, LoI only, moved to Green Book.
- Green Book readiness: missing activities, missing funding allocation, missing disbursement plan, stale BB snapshot.
- DK bottleneck: age since DK date, no LA, final status, financing-detail quality.
- LA health: not effective, near closing, extended, missing cumulative disbursement.

**Impact**

- Dashboard answers "where is the portfolio distributed?" better than "what must be acted on next?"
- Bottlenecks are visible as a count but not actionable as a queue.

**Improve**

- Add an "Action Queue" panel per stage.
- Keep distribution as secondary tabs.
- Provide drilldown links to filtered Project Master or module pages.

### Finding 4 - Permission model is too broad

**Status: accepted temporary policy.**

The route and sidebar gate dashboard access with `bb_project` read permission. The dashboard presents data that spans Green Book, Daftar Kegiatan, Loan Agreement, lenders, institutions, regions, and possibly monitoring.

**Impact**

- A STAFF user with only Blue Book access may see aggregated downstream data they cannot otherwise access.
- Conversely, a user with LA/DK access but without BB access cannot access the dashboard.

**Improve**

- Define dashboard permission policy explicitly:
  - Option A: require a dedicated `dashboard.read` permission.
  - Option B: show only sections allowed by module permissions.
  - Option C: require all relevant read permissions for a full dashboard.
- Document the chosen behavior in API and frontend route metadata.

### Finding 5 - Dashboard contract is not durable yet

There are generated backend files named `dashboard.sql.go` and `dashboard_analytics.sql.go`, but no matching source SQL files under `prism-backend/sql/queries/`. There is also no current frontend dashboard service/store.

**Impact**

- Running `sqlc generate` can leave stale generated files or remove old dashboard functionality unexpectedly.
- Future agents may trust generated query files that no longer have source ownership.
- API contract does not define a dashboard endpoint for the current `/dashboard` page.

**Improve**

- Either restore dashboard SQL source and service/handler/route wiring, or remove stale generated dashboard query files.
- Add dashboard endpoint documentation to `docs/PRISM_API_Contract.md`.
- Define response shape for KPI, stage totals, stage conversion, and drilldown filters.

### Finding 6 - Component typing should become discriminated by panel kind

`DashboardStagePanel` allows `data`, `rows`, and `pairs` to be optional for every panel kind, while `PlanningFunnelFlow.vue` assumes different fields based on `panel.kind`.

**Impact**

- A malformed panel can silently render empty or break at runtime.
- Adding new panel types will increase template branching and weak typing.

**Improve**

- Change `DashboardStagePanel` into a discriminated union:
  - `bars`, `donut`, `donutbar`, `regions`, `stack`, `status`, `card` require `data`.
  - `table` requires `rows`.
  - `flow` requires `pairs`.
- Split panel renderers into small presentational components once data is live.

### Finding 7 - Accessibility role should not wrap interactive content

**Status: implemented in frontend.**

`PlanningFunnelFlow.vue` sets `role="img"` on a section that contains clickable stage buttons and tabs.

**Impact**

- Assistive technology may treat the section as one image instead of a group of interactive controls.
- Button and tab semantics can become harder to navigate.

**Improve**

- Use `role="region"` or normal section semantics with `aria-labelledby`.
- Keep the chart summary as text, not as the role of the whole interactive container.
- Use actual `button` controls for stage expand/collapse and a proper tabs pattern for stage detail tabs.

---

## 4. Recommended Dashboard Information Architecture

Use one `/dashboard` page with a primary planning funnel and operational tabs:

1. `Ringkasan`
   - Total project by stage.
   - Total value by stage.
   - Conversion by transition.
   - Top bottleneck stage.

2. `Blue Book Readiness`
   - Lender indication coverage.
   - LoI coverage.
   - Projects not yet in Green Book.
   - Institution, program, and region distribution.

3. `Green Book Readiness`
   - BB -> GB relationship pattern.
   - Funding source completeness.
   - Activity and funding allocation completeness.
   - Disbursement plan completeness.
   - Projects not yet in DK.

4. `Daftar Kegiatan Queue`
   - DK Projects without LA.
   - Queue age buckets.
   - Fixed lender distribution.
   - Financing detail and loan allocation completeness.

5. `Loan Agreement Health`
   - Signed, effective, not effective.
   - Near closing.
   - Extended.
   - DK Projects with multiple LA.

6. `Data Quality`
   - Missing lender.
   - Missing region.
   - Missing institution.
   - Missing amount.
   - Stale snapshot indicators across revisions.

Monitoring Disbursement should be a separate tab only if the product decision is to include execution monitoring in `/dashboard`.

---

## 5. Proposed Data Contract

Recommended minimum response for a planning funnel endpoint:

```json
{
  "data": {
    "last_synced_at": "2026-05-06T02:14:00Z",
    "filters": {
      "period_id": null,
      "publish_year": null,
      "include_history": false
    },
    "kpis": [
      {
        "key": "active_projects",
        "label": "Total Proyek Aktif",
        "value": 167,
        "unit": "proyek",
        "tone": "neutral"
      }
    ],
    "stages": [
      {
        "key": "BB",
        "label": "Blue Book",
        "project_count": 96,
        "amount_usd": 29280000000,
        "pipeline_share_pct": 100,
        "next_stage": "GB",
        "next_conversion_pct": 37.5,
        "blocked_count": 60
      }
    ],
    "action_queues": [
      {
        "stage": "DK",
        "risk_code": "DK_WITHOUT_LA",
        "label": "DK belum menjadi Loan Agreement",
        "project_count": 35,
        "oldest_age_days": 187
      }
    ]
  }
}
```

Recommended query filters:

- `period_id`
- `publish_year`
- `include_history`
- `institution_ids`
- `lender_ids`
- `lender_types`
- `program_title_ids`
- `region_ids`
- `stage`
- `min_age_days`
- `search`

---

## 6. Improvement Backlog

### P0 - Required before operational use

- Replace hard-coded KPI and funnel numbers with API-backed data.
- Resolve dashboard scope: planning/legal only or planning plus monitoring.
- Define dashboard permission policy.
- Add dashboard API contract and backend source queries.
- Remove contradictory copy: `real-time` vs `Data contoh`.

### P1 - Required for business usefulness

- Add action queue per transition:
  - BB without lender indication.
  - BB with LoI but no GB.
  - GB with incomplete readiness.
  - DK without LA.
  - LA signed but not effective.
  - LA near closing or extended.
- Add drilldown links to filtered module pages.
- Add filter controls for period, publish year, K/L, lender, program, region, and include history.
- Show latest vs historical snapshot behavior explicitly.

### P2 - UI and technical quality

- Convert panel types into discriminated TypeScript unions.
- Split panel rendering into small components.
- Replace `role="img"` with accessible region and tab semantics.
- Avoid inline grid style for `donutbar`; use responsive classes or CSS.
- Add loading, empty, and error states.
- Add export only after real filter/data contract exists.

---

## 7. Suggested Implementation Order

1. Decide scope: keep Monitoring outside `/dashboard` or add it as explicit fifth stage.
2. Define dashboard API contract in `docs/PRISM_API_Contract.md`.
3. Add backend SQL source under `prism-backend/sql/queries/dashboard.sql`.
4. Run `make generate`.
5. Add dashboard model, service, handler, and route with permission middleware.
6. Add frontend `dashboard.types.ts`, `dashboard.service.ts`, and `dashboard.store.ts`.
7. Replace hard-coded constants in `DashboardPage.vue`.
8. Add action queues and drilldown links.
9. Refactor `PlanningFunnelFlow.vue` panel rendering after data is stable.
10. Verify with:
    - `go test ./...`
    - `npm.cmd run type-check`
    - `npm.cmd run build`
    - `git diff --check`
    - Browser smoke test at `/dashboard`
