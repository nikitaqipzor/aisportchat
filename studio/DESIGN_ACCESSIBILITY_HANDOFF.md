# Design and accessibility handoff

## Goal
Bring the primary workout choice above secondary Home cards; improve navigation legibility, muscle selection touch targets, and exercise instruction hierarchy.

## Changes
- Home prioritizes a manual workout action, environment choice, and muscle map before supplementary recovery/device/coach/program cards. The manual action still requires a verified environment. Localized several Home captions and replaced a false fixed-name profile initial with «Я».
- BottomNavigation uses descriptive 12sp labels, allows wrapping for large text, and marks the active tab visually and through selected accessibility state.
- BodyMap keeps a compact noninteractive front/back silhouette, hidden from TalkBack and unable to intercept touch. Front/back region groups use full 48dp independent muscle buttons instead of overlapping fixed-position hotspots. All 11 catalog muscle groups remain selectable, including the posterior options.
- ExerciseGuide highlights the first supplied technique tip, renders source instructions as separate numbered cards, and provides an accessible 48dp close button. No generated exercise instructions or medical claims were added.

## Risks
- Code/layout checks cannot prove actual line wrapping or TalkBack focus order on a physical Android device. Verify 320dp viewport, 200% font scaling, both body sides, and screen reader announcements.
- The BodyMap silhouette is a simplified navigation cue, not a detailed anatomical illustration. ExerciseGuide still requires curated photos/video or professionally checked diagrams for complete visual form instruction.

## Tests/evidence
- `node scripts/test-mobile-design-accessibility.mjs` PASS: primary action order, selection coverage, touch and accessibility contracts.
- `node scripts/test-mobile-home-reliability.mjs` PASS: environment/profile safety retained.
- `node scripts/test-mobile-core-screens.mjs` PASS.
- `npm run typecheck --prefix apps/mobile` PASS after parallel edit settled.
- `git diff --check` PASS at handoff.

## Unresolved
Run on a physical Android device with TalkBack and enlarged system font before considering accessibility complete. Validate bottom navigation at 320dp and group labels in both orientations and languages as available.

## Handoff
Lead engineer: integrate these files alongside independent screen fixes, repeat mobile TypeScript typecheck, then request Android QA of the touch and screen-reader flows.
