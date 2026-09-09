# Static Mobile CI Fix Handoff

## Goal

Restore the `static-mobile` release-gate job by making its TypeScript verifier independent of a specific Node installation path.

## Changes

- Updated `scripts/verify-mobile-syntax.mjs` to load TypeScript from the local package when available, then fall back to the active global npm root.
- Removed the hard-coded `/opt/nvm/versions/node/v22.16.0/...` import, which did not match the TypeScript installation selected by `actions/setup-node@v4` and `npm install --global`.

## Risks

- Low: the loading strategy matches the existing mobile regression scripts.
- The verifier still requires TypeScript to be installed either locally or globally; the CI job already installs `typescript@6.0.3` globally.

## Tests / Evidence

Failure reproduced before the change:

```text
Error [ERR_MODULE_NOT_FOUND]: Cannot find module '/opt/nvm/versions/node/v22.16.0/lib/node_modules/typescript/lib/typescript.js'
```

Executed with `typescript@6.0.3` installed into an isolated temporary global npm prefix:

```text
mobile syntax: 58 files, 0 error(s)
mobile auth refresh: PASS (refreshCalls=1, protectedCalls=4)
mobile session scope: PASS
mobile technique mapping: PASS
mobile passive health sync owner isolation: PASS
```

## Unresolved Items

- None within the `static-mobile` job scope.
- Backend, Android CI, and screen implementation were intentionally not changed or evaluated.

## Handoff

Lead/QA: review the two-file diff and rerun the `static-mobile` GitHub Actions job on the integrated branch.
