# Push AI Fitness OS Sprint 4C R2 to aisportchat

This directory is the unpacked Sprint 4C R2 source, without an extra wrapper folder.

## One-time push

```bash
cd aisportchat-r2-ready
git init
git branch -M recovery/r2-clean-restart-2
git remote add origin https://github.com/nikitaqipzor/aisportchat.git
git add .
git commit -m "recovery: import AI Fitness OS Sprint 4C R2"
git push -u origin recovery/r2-clean-restart-2 --force
```

After push, verify that these paths exist in GitHub:

- apps/mobile
- services/api
- packages/contracts
- infra
- scripts
- docs
- .github/workflows/ci.yml

## Local checks after clone

```bash
./scripts/verify.sh
./scripts/release-preflight.sh
```
