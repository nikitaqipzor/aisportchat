#!/usr/bin/env python3
from pathlib import Path
import re
import sys

root = Path(__file__).resolve().parents[1]
migrations = root / "services/api/migrations"
runner = (migrations / "run.sh").read_text()
smoke = (root / "scripts/postgres-migration-smoke.sh").read_text()
pattern = re.compile(r"^(\d{6})_(.+)\.(up|down)\.sql$")
pairs = {}
errors = []

if "pg_advisory_xact_lock(20260928, 23001)" not in runner:
    errors.append("migration runner must serialize ledger creation and DDL with the shared advisory lock")
if "already_applied \\\\gset" not in runner or "\\if :already_applied" not in runner:
    errors.append("migration runner must recheck the ledger under the lock before applying DDL")
if "ALLOW_DESTRUCTIVE_MIGRATION_SMOKE" not in smoke:
    errors.append("PostgreSQL migration smoke must require explicit disposable-database opt-in")
if smoke.count('"$MIGRATIONS/run.sh" &') < 2:
    errors.append("PostgreSQL migration smoke must exercise concurrent migration runners")

for path in migrations.iterdir():
    if not path.is_file() or path.name == "run.sh":
        continue
    match = pattern.match(path.name)
    if not match:
        errors.append(f"unexpected migration filename: {path.name}")
        continue
    number, name, direction = match.groups()
    key = (number, name)
    pairs.setdefault(key, set()).add(direction)

for key, directions in sorted(pairs.items()):
    if directions != {"up", "down"}:
        errors.append(f"migration {key[0]}_{key[1]} missing pair: have {sorted(directions)}")

numbers = {}
for number, name in pairs:
    numbers.setdefault(number, set()).add(name)
for number, names in sorted(numbers.items()):
    if len(names) != 1:
        errors.append(f"migration number {number} reused by {sorted(names)}")

ordered = sorted(int(number) for number in numbers)
if ordered:
    expected = list(range(ordered[0], ordered[-1] + 1))
    if ordered != expected:
        errors.append(f"migration sequence has gaps: got {ordered}, expected {expected}")

if errors:
    for error in errors:
        print("FAIL", error)
    sys.exit(1)

print(f"Migration pairs aligned: {len(pairs)} up/down pairs, sequence {ordered[0]:06d}-{ordered[-1]:06d}")
