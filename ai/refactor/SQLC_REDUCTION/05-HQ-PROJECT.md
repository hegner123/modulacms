# HQ Multi-Agent Project Configuration

This document contains the HQ project configuration for multi-agent parallel execution.

---

## Project Configuration JSON

```json
{
  "name": "sqlc-type-unification",
  "base_commit": "<current HEAD>",
  "steps": [
    {"step_num": 0, "branch": "baseline/validate-current", "scope": "validation", "depends_on": []},
    {"step_num": 1, "branch": "schema/change-events-table", "scope": "schema", "depends_on": [0]},
    {"step_num": 2, "branch": "schema/backup-tables", "scope": "schema", "depends_on": [0]},
    {"step_num": 3, "branch": "schema/fk-indexes", "scope": "schema", "depends_on": [0]},
    {"step_num": 4, "branch": "schema/check-constraints", "scope": "schema", "depends_on": [0]},
    {"step_num": 5, "branch": "foundation/feature-branch", "scope": "setup", "depends_on": [1, 2, 3, 4]},
    {"step_num": 6, "branch": "types/primary-ids-ulid", "scope": "types", "depends_on": [5]},
    {"step_num": 7, "branch": "types/nullable-ids", "scope": "types", "depends_on": [5]},
    {"step_num": 8, "branch": "types/timestamp-hlc", "scope": "types", "depends_on": [5]},
    {"step_num": 9, "branch": "types/enums-distributed", "scope": "types", "depends_on": [5]},
    {"step_num": 10, "branch": "types/validation", "scope": "types", "depends_on": [5]},
    {"step_num": 11, "branch": "types/change-events-hlc", "scope": "types", "depends_on": [5]},
    {"step_num": 12, "branch": "types/transaction", "scope": "types", "depends_on": [5]},
    {"step_num": 13, "branch": "config/sqlc-yml", "scope": "config", "depends_on": [6, 7, 8, 9, 10, 11]},
    {"step_num": 14, "branch": "validate/circular-imports", "scope": "validation", "depends_on": [13]},
    {"step_num": 15, "branch": "generate/sqlc-regen", "scope": "codegen", "depends_on": [14]},
    {"step_num": 16, "branch": "validate/type-check", "scope": "validation", "depends_on": [15]},
    {"step_num": 171, "branch": "simplify/struct-definitions", "scope": "wrapper", "depends_on": [16]},
    {"step_num": 172, "branch": "simplify/remove-json-structs", "scope": "wrapper", "depends_on": [171]},
    {"step_num": 173, "branch": "simplify/remove-formparams", "scope": "wrapper", "depends_on": [171]},
    {"step_num": 18, "branch": "simplify/permission", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 19, "branch": "simplify/role", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 20, "branch": "simplify/media-dimension", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 21, "branch": "simplify/user", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 22, "branch": "simplify/admin-route", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 23, "branch": "simplify/route", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 24, "branch": "simplify/datatype", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 25, "branch": "simplify/field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 26, "branch": "simplify/admin-datatype", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 27, "branch": "simplify/admin-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 28, "branch": "simplify/token", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 29, "branch": "simplify/user-oauth", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 30, "branch": "simplify/table", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 31, "branch": "simplify/media", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 32, "branch": "simplify/session", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 33, "branch": "simplify/content-data", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 34, "branch": "simplify/content-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 35, "branch": "simplify/admin-content-data", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 36, "branch": "simplify/admin-content-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 37, "branch": "simplify/datatype-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 38, "branch": "simplify/admin-datatype-field", "scope": "wrapper", "depends_on": [172, 173]},
    {"step_num": 39, "branch": "cleanup/dead-code", "scope": "cleanup", "depends_on": [18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38]},
    {"step_num": 40, "branch": "integration/api-handlers", "scope": "integration", "depends_on": [39]},
    {"step_num": 41, "branch": "integration/cli-ops", "scope": "integration", "depends_on": [39]},
    {"step_num": 42, "branch": "testing/test-suite", "scope": "testing", "depends_on": [40, 41]},
    {"step_num": 43, "branch": "testing/type-validation", "scope": "testing", "depends_on": [42]},
    {"step_num": 44, "branch": "docs/interface-docs", "scope": "docs", "depends_on": [43]},
    {"step_num": 45, "branch": "docs/final-docs", "scope": "docs", "depends_on": [44]}
  ]
}
```

---

## Step Numbering Notes

- Steps 1-4 are schema improvements (Phase 0.5)
- Steps 171, 172, 173 are "sub-steps" of conceptual step 17 (17a, 17b, 17c in the plan)
- Schema steps 3 and 4 can run in parallel after step 0

---

## Dependency Graph

```
Timeline:

Step 0 ─┬─ 1 (change_events) ─┬─ 5 ─┬─ 6 ─┬──────────────────────────────────────────────────────────────┐
        ├─ 2 (backups) ───────┤     ├─ 7 ─┤                                                               │
        ├─ 3 (fk-indexes) ────┤     ├─ 8 ─┤                                                               │
        └─ 4 (constraints) ───┘     ├─ 9 ─┼─ 13 ─ 14 ─ 15 ─ 16 ─ 17a ─┬─ 17b ─┬─ Steps 18-38 (21) ─┬─ 39 ─┬─ 40 ─┬─ 42 ─ 43 ─ 44 ─ 45
                                    ├─ 10 ┤                            └─ 17c ─┘                    │      └─ 41 ─┘
                                    ├─ 11 ┤                                                         │
                                    └─ 12 ┘                                                         │

Phase 0 (Baseline):      1 step (must complete first)
Phase 0.5 (Schema):      4 steps (all parallel after Step 0 - change_events, backups, fk-indexes, constraints)
Phase 1 (Types):         7 steps (parallel after step 5)
Phase 2 (Config):        4 sequential steps (13, 14, 15, 16)
Phase 3 (Wrappers):      24 steps (17a → 17b||17c → 21 parallel)
Phase 4 (Cleanup):       7 steps (some parallel)
Total:                   49 HQ steps (including sub-steps and schema)
```

**Critical Path:** 0 → 1 → 5 → types → 13 → 14 → 15 → 16 → 17a → 17b/c → wrappers → 39 → integration → tests → docs

---

## Scope Categories

| Scope | Description | Steps |
|-------|-------------|-------|
| validation | Verify current state, check for issues | 0, 14, 16 |
| schema | Database schema changes | 1, 2, 3, 4 |
| setup | Branch creation, environment setup | 5 |
| types | Custom Go type definitions | 6, 7, 8, 9, 10, 11, 12 |
| config | Configuration file changes | 13 |
| codegen | Code generation (sqlc) | 15 |
| wrapper | Wrapper layer simplification | 171, 172, 173, 18-38 |
| cleanup | Dead code removal | 39 |
| integration | API and CLI updates | 40, 41 |
| testing | Test suite execution | 42, 43 |
| docs | Documentation updates | 44, 45 |

---

## Launching the Project

```bash
# Get current commit hash
BASE_COMMIT=$(git rev-parse HEAD)

# Create project using HQ MCP tool
# Replace <current HEAD> with actual commit hash in the JSON
```

Or use the MCP tool directly:

```
mcp__hq__create_project with the JSON above
```

---

## Monitoring Progress

```bash
# Get project status
mcp__hq__get_project {"name": "sqlc-type-unification"}

# Get available steps (ready to claim)
mcp__hq__get_available_steps {"project": "sqlc-type-unification"}

# Get metrics
mcp__hq__get_metrics {"project": "sqlc-type-unification"}

# Check for stale work
mcp__hq__detect_stale_work {"timeout_minutes": 15}
```

---

## Agent Coordination

### Claiming Steps

Each agent should:
1. Claim an available step: `mcp__hq__claim_step {"project": "...", "agent_id": "agent-N"}`
2. Start the step: `mcp__hq__start_step {"step_id": X}`
3. Send heartbeats every 30-60s: `mcp__hq__heartbeat {"step_id": X, "agent_id": "agent-N"}`
4. Complete: `mcp__hq__complete_step {"step_id": X, "commit_hash": "...", "files_modified": [...], "notes": "..."}`

### Recovery

If an agent crashes or times out:
1. Detect stale work: `mcp__hq__detect_stale_work {"timeout_minutes": 15}`
2. Recover the step: `mcp__hq__recover_step {"step_id": X}`
3. New agent can claim it

### Best Practices

- Use consistent agent IDs (e.g., "agent-1", "agent-2")
- Send heartbeats regularly while working
- Complete steps promptly after finishing
- Include meaningful commit hashes and file lists
- Add notes about any issues encountered
