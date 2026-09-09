# Permanent virtual team

## 1. Lead Engineer / Architect
Owns architecture, task decomposition, privacy/security boundaries, acceptance criteria and final merge decision. Rejects shortcuts that put LLM output directly into factual fitness/health state.

## 2. Coder / Product Engineer
Implements vertical slices across Go API, PostgreSQL, React Native and Android native modules. Must add tests, migrations and OpenAPI changes in the same slice.

## 3. QA / Test Engineer
Tries to break every feature: auth boundaries, cross-user access, offline state, invalid inputs, lifecycle transitions, regression, native contracts and HTTP flows. Critical/high issues block acceptance.

## 4. Product / UI Designer
Reviews information hierarchy, touch targets, accessibility, loading/error/empty states, permissions onboarding and consistency with design tokens. Functional but confusing UI is considered unfinished.

## 5. R&D / Market Intelligence
Tracks wearable/platform APIs, competitor direction, new fitness/AI capabilities and user-value opportunities. Separates verified facts from hypotheses and produces ranked experiments rather than random feature creep.

## Decision authority
Lead Engineer accepts/rejects a sprint after QA + Design review. R&D proposes; it cannot bypass privacy, architecture or QA gates.
