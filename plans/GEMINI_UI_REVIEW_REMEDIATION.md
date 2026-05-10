# Gemini Delegation Instructions: PRISM UI Consistency and Reuse Remediation

## Objective

Fix UI inconsistencies, duplicated UI logic, and redundant asset usage found in the PRISM frontend review. Work only inside `prism-frontend/` unless this file explicitly says a docs update is required.

This is a remediation task, not a redesign. Preserve current product behavior, API contracts, permissions, and business rules.

## Required Reading Before Editing

Read these files first, in this order:

1. `AGENTS.md`
2. `docs/PRISM_Business_Rules.md`
3. `docs/PRISM_Frontend_Structure.md`
4. `docs/PRISM_API_Contract.md` only if a UI change touches API request/response assumptions

## Hard Constraints

- Vue code must use Vue 3 Composition API and `<script setup lang="ts">`.
- Do not introduce `any` unless there is no safer alternative, and explain why in the final report.
- Do not call Axios directly from pages/components. API calls stay in `src/services/`, state in Pinia stores.
- Do not create or edit `tailwind.config.ts` or `postcss.config.ts`.
- Tailwind v4 styling must stay in `src/assets/styles/main.css` via `@theme`.
- PrimeVue component styling should use theme tokens, Pass-Through API, or existing component props. Do not add `!important` overrides.
- Do not add new dependencies without explicit approval.
- Do not change backend behavior for this task.
- Do not remove existing user-facing functionality while extracting reusable components.

## Primary Findings To Fix

### 1. P1: Mobile navigation unavailable after login

Evidence:

- `prism-frontend/src/layouts/components/AppSidebar.vue:216-218`
- `prism-frontend/src/layouts/components/AppTopbar.vue:9-17`
- Browser smoke at `/projects` on a tablet/mobile width shows the sidebar is hidden and no menu trigger exists.

Expected fix:

- Add a mobile/tablet navigation path for authenticated layout.
- Preserve desktop sidebar behavior on `lg` and wider.
- Add a menu button in `AppTopbar` for smaller viewports.
- Implement either a reusable drawer component or a responsive sidebar overlay.
- Ensure overlay closes after route navigation and when user clicks outside or presses Escape.
- Keep permission-based menu filtering intact.

Suggested implementation shape:

- `AppLayout.vue` owns `isMobileSidebarOpen`.
- `AppTopbar.vue` emits `toggle-sidebar`.
- `AppSidebar.vue` accepts responsive state/close callback or a wrapper drawer controls it.
- Reuse the same navigation item data instead of duplicating menu definitions.

Acceptance criteria:

- At mobile/tablet width, an authenticated user can open navigation, navigate to another page, and close the menu.
- At desktop width, existing sidebar behavior remains unchanged.
- Browser smoke covers `/login` -> login -> `/projects` and at least one navigation click on narrow viewport.

### 2. P2: MasterTreeTable uses raw Paginator instead of reusable footer

Evidence:

- `prism-frontend/src/pages/master/MasterTreeTable.vue:118-124`
- Shared component exists at `prism-frontend/src/components/common/ListPaginationFooter.vue`.

Expected fix:

- Replace direct `Paginator` usage with `ListPaginationFooter` or extract a compatible shared pagination wrapper if the tree table needs a slightly different contract.
- Preserve server-side root pagination and lazy child loading behavior.
- Ensure changing rows per page resets to page 1 consistently.
- Keep display text consistent with other list pages.

Acceptance criteria:

- Institution, Region, Program Title, and Bappenas Partner tree pages show the same pagination footer style and behavior as other list pages.
- Paging root rows does not break expanded child loading.

### 3. P2: Wide editable tables are clipped on small screens

Evidence:

- `prism-frontend/src/components/daftar-kegiatan/FinancingDetailTable.vue:50-51`
- `prism-frontend/src/components/daftar-kegiatan/LoanAllocationTable.vue:45-46`
- `prism-frontend/src/components/green-book/ActivitiesTable.vue:42-43`
- `prism-frontend/src/components/blue-book/LenderIndicationTable.vue:49-50`
- `prism-frontend/src/components/blue-book/LoITable.vue:16`

Expected fix:

- Replace clipping wrappers with horizontal scroll behavior.
- Prefer a reusable `EditableTableShell` if it can be introduced cleanly without over-refactoring.
- Preserve current validation messages, add/remove row actions, empty states, and totals.

Acceptance criteria:

- All affected tables remain usable at mobile/tablet widths.
- No action column or rightmost input is unreachable.
- Existing desktop layout is not visually degraded.

### 4. P2: Status labels are not centralized

Evidence:

- `prism-frontend/src/components/common/StatusBadge.vue:23-24`
- Some pages pass labels manually, while others display internal values such as `active`, `deleted`, or `extended`.

Expected fix:

- Centralize status label and severity mapping.
- Preserve the current explicit `label` prop as an override.
- Avoid breaking Blue Book and Green Book status labels that are already business-rule sensitive.

Suggested implementation shape:

- Add `src/utils/status-labels.ts` or equivalent.
- Support domain-aware mappings where needed, for example `user`, `blue_book`, `green_book`, `loan_agreement`, `import`.
- Update pages that currently leak internal status values.

Acceptance criteria:

- User-facing status labels are consistent and localized.
- No page displays raw internal status values unless intentionally required by business wording.

## Secondary Findings To Address If Scope Allows

### 5. Master list search and filter duplication

Evidence examples:

- `prism-frontend/src/pages/master/CountryPage.vue:103-111`
- `prism-frontend/src/pages/master/LenderPage.vue:138-158`
- `prism-frontend/src/pages/master/CurrencyPage.vue:137-145`
- `prism-frontend/src/pages/master/NationalPriorityPage.vue:122-145`

Expected direction:

- Align master list pages with `SearchFilterBar.vue` and `useListControls.ts`, or create a master-specific wrapper that preserves the same visual and interaction contract.
- Do not regress existing search, filter, sorting, and pagination behavior.

### 6. Currency, number, and date formatting duplication

Evidence examples:

- `prism-frontend/src/components/common/CurrencyDisplay.vue`
- `prism-frontend/src/components/project/SummaryCard.vue`
- `prism-frontend/src/pages/ProjectMasterPage.vue`
- `prism-frontend/src/pages/SpatialDistributionPage.vue`
- Repeated `formatDate` and `toFormErrors` helpers in page utility files.

Expected direction:

- Introduce shared formatter utilities, for example `src/utils/formatters.ts`.
- Introduce shared form error normalization, for example `src/utils/form-errors.ts`.
- Update call sites incrementally and preserve existing display precision.

### 7. Large duplicated dropdown overlay logic

Evidence examples:

- `prism-frontend/src/components/common/MultiSelectDropdown.vue`
- `prism-frontend/src/components/common/SingleSelectDropdown.vue`

Expected direction:

- Extract shared option resolution, filtering, floating overlay position, outside-click handling, and keyboard behavior into a composable or base component.
- Keep the single-select and multi-select public APIs backward compatible.

### 8. DK financing and loan allocation table duplication

Evidence examples:

- `prism-frontend/src/components/daftar-kegiatan/FinancingDetailTable.vue`
- `prism-frontend/src/components/daftar-kegiatan/LoanAllocationTable.vue`

Expected direction:

- Extract a shared multi-currency row/table helper or component.
- Keep first-column differences slot-based: lender selector for financing, institution selector for allocation.

### 9. Large page components should be split

High-risk files:

- `prism-frontend/src/pages/blue-book/BlueBookDetailPage.vue`
- `prism-frontend/src/pages/green-book/GreenBookDetailPage.vue`
- `prism-frontend/src/pages/SpatialDistributionPage.vue`
- `prism-frontend/src/pages/ImportDataPage.vue`
- `prism-frontend/src/pages/ProjectMasterPage.vue`

Expected direction:

- Extract only cohesive, low-risk parts.
- Prioritize components reused in multiple pages, such as import summary, revision import dialog, project filters, project table, spatial summary cards, and detail section cards.
- Do not perform broad rewrites unless necessary for the primary findings.

## Asset Redundancy To Fix

### Duplicate logo/favicon asset

Evidence:

- `prism-frontend/public/favicon.png`
- `prism-frontend/public/prism-logo.png`
- The two PNG files are byte-identical 512x512 images.
- `prism-frontend/index.html` references both.
- `AppSidebar.vue` and `LoginPage.vue` reference `/prism-logo.png`.

Expected fix:

- Keep a single canonical logo source.
- Generate or keep size-appropriate derivatives only if required:
  - favicon: 32x32 or 64x64
  - apple touch icon: 180x180
  - app/sidebar logo: optimized SVG, WebP, or compressed PNG
- Update references consistently.
- Do not remove `favicon.ico` unless verified unused and safe.

Acceptance criteria:

- Production build no longer contains two byte-identical large logo PNGs.
- Browser favicon and login/sidebar logo still render correctly.

### Large map assets

Evidence:

- `prism-frontend/public/maps` is large and copied to production.
- `SpatialChoroplethMap.vue` lazy-loads map JSON at runtime.

Expected direction:

- Do not remove map files unless spatial features are verified.
- If optimizing, prefer simplification/compression or moving to versioned static hosting.
- Treat this as lower priority than duplicate logo cleanup.

## Monitoring UI Gap

The frontend structure docs list monitoring pages and components, but the current source appears to contain only `src/components/monitoring/AbsorptionBar.vue`.

Do not implement the full monitoring module as part of this remediation unless explicitly requested. Instead:

- Confirm whether monitoring UI is intentionally out of scope on the current branch.
- If it is out of scope, update the final report with the gap and recommended follow-up.
- If instructed to implement, follow the frontend feature order from `AGENTS.md`: API contract -> types -> schema -> service -> store -> composable -> components -> pages -> routes.

## Verification Required

Run these checks after implementation:

1. `cd prism-frontend && npm run type-check`
2. `cd prism-frontend && npm run build-only`
3. `cd prism-frontend && npm run lint`

Browser smoke required:

1. Open `/login`.
2. Login with available local seed credentials.
3. Verify `/projects` desktop layout.
4. Verify `/projects` mobile/tablet navigation.
5. Verify at least one master tree page pagination.
6. Verify at least one affected editable table on narrow viewport.
7. Verify favicon and sidebar/login logo render after asset cleanup.

If any command cannot run because of local environment limits, report the exact blocker and the strongest verification completed.

## Final Report Format For Gemini

Return a concise report with these sections:

1. Summary of changes
2. Files changed
3. Findings fixed, mapped to the finding numbers above
4. Verification results with exact command outcomes
5. Known gaps or follow-up tasks

Do not claim a finding is fixed unless it was verified.
