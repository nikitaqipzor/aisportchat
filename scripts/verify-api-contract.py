#!/usr/bin/env python3
from pathlib import Path
import re
import sys
import yaml

root = Path(__file__).resolve().parents[1]
router_text = (root / "services/api/internal/httpapi/router.go").read_text()
spec = yaml.safe_load((root / "services/api/openapi/openapi.yaml").read_text())

methods = {"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
route_re = re.compile(r'mux\.Handle(?:Func)?\("(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS) ([^"]+)"')
routes = set(route_re.findall(router_text))
normalized_routes = {
    (method, path[len("/api/v1") :] if path.startswith("/api/v1") else path)
    for method, path in routes
}

spec_ops = {
    (method.upper(), path)
    for path, item in spec.get("paths", {}).items()
    for method in item
    if method.upper() in methods
}

# /healthz is intentionally an infrastructure endpoint outside the public /api/v1 spec.
allowed_router_only = {("GET", "/healthz")}
router_only = normalized_routes - spec_ops - allowed_router_only
spec_only = spec_ops - normalized_routes

if router_only or spec_only:
    if router_only:
        print("Router operations missing from OpenAPI:")
        for op in sorted(router_only):
            print("  ", op)
    if spec_only:
        print("OpenAPI operations missing from router:")
        for op in sorted(spec_only):
            print("  ", op)
    sys.exit(1)

print(f"API contract aligned: {len(spec_ops)} OpenAPI operations, {len(routes)} router routes (+ /healthz)")
