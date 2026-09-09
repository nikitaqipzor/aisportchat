# Backend custom-product CI fix

## Goal

Restore the PostgreSQL integration scenario for creating a user-owned custom
food without changing the production API contract or weakening strict request
decoding.

## Changes and findings

- Updated the custom-food payload in
  `services/api/internal/httpapi/postgres_integration_test.go` to use the
  canonical request fields `kcal_per_100g` and `serving_g`.
- Added the contract-required `fiber_per_100g` field to the integration
  fixture.
- Root cause: this integration fixture still used the obsolete names
  `calories_per_100g` and `serving_size_g`. `decodeJSON` intentionally calls
  `DisallowUnknownFields`, so the first obsolete property caused HTTP 400 with
  `invalid request body` before service or PostgreSQL code ran.
- The handler input type, OpenAPI `CustomFoodInput`, mobile API client, and the
  existing memory-backed HTTP test already agree on the canonical field names;
  no production handler or contract change was needed.

## Risks

- No runtime behavior changed. The fix only aligns a PostgreSQL integration
  fixture with the existing API contract.
- Full execution still depends on Go 1.23 and the PostgreSQL integration-test
  environment used by CI.

## Tests and evidence

- Static contract comparison completed across:
  - `nutrition.CustomFoodInput` JSON tags;
  - OpenAPI `CustomFoodInput`;
  - the mobile `createCustomFood` payload type;
  - the passing memory-backed HTTP custom-food scenario.
- Confirmed the failing fixture no longer contains the obsolete property names.
- Go tests were not executed locally because this environment does not provide
  a Go executable. CI must run the PostgreSQL integration test.

## Unresolved items

- Re-run the backend/PostgreSQL CI job to confirm the corrected fixture against
  the provisioned database.

## Handoff

QA should run the backend test suite, including the PostgreSQL integration
scenario, and confirm that custom-food creation returns HTTP 201 and subsequent
entry aggregation and cross-user privacy assertions pass.
