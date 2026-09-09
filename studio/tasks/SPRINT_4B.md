# Sprint 4B task board — Personal Baselines & Provenance

## Lead Engineer
- [x] define baseline leakage rule (exclude current day)
- [x] define source conflict policy
- [x] define freshness policy and stale fallback
- [x] prevent activity double-counting in readiness

## Coder
- [x] health insights service + HTTP API
- [x] 7/28-day baselines + deviations
- [x] source-aware storage migration
- [x] deterministic multi-source resolver
- [x] stale wearable fallback in Recovery
- [x] AI Coach read-only health insights tool
- [x] Connected Devices baseline/provenance UI
- [x] Recovery baseline/freshness UI

## QA
- [x] baseline excludes current day
- [x] empty days do not inflate coverage
- [x] source resolution test
- [x] stale fallback test
- [x] HTTP insights test
- [x] full Go/mobile/native regression gate

## Designer
- [x] confidence/freshness hierarchy
- [x] 28-day sync progress state
- [x] source conflict is visible instead of hidden
- [x] recovery explains wearable vs manual source

## R&D
- [x] rank next experiments
- [x] keep direct vendor API below aggregator-first priority
- [x] propose readiness explanation timeline as next differentiation slice
