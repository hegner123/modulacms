# Team Memory - Quick Reference

## Setup (One-Time)
```bash
cd ai/mcp
npm install && npm run build
```
Enable `team-memory` server when Claude Code prompts you.

## Common Commands

### Search First
```
Search memories for "GSAP animation"
```

### Record Decision
```
Record design decision: "Use Server Components by default" with rationale "Better performance, smaller bundles" and made_by "Your Name"
```

### Add Best Practice
```
Add best practice: title "Clean up GSAP timelines" and description "Always revert context in useLayoutEffect cleanup" and category "frontend" and added_by "Your Name"
```

### Log Issue
```
Record known issue: title "Issue name" and description "Details here" and workaround "How to work around it" and severity "medium" and reported_by "Your Name"
```

### Track Client Customization
```
Record client customization for client_name "ClientName" with customization_type "theme" and description "What was changed" and shared_vs_custom "custom"
```

### Agent Workflow
```
Store component contract for "ComponentName" with contract_json "{...}" and validation_status "passed"

Log agent execution: agent_name "backend-agent" with task_description "Generated UDA files" and success true and executed_by "AI Agent"
```

### Get Recent Activity
```
Get recent activity from the last 7 days

Get component history for "HeroSection"
```

### List Issues
```
List open issues

List open issues with severity "high"
```

## Categories

**Memory categories**: `design-decision`, `architecture`, `component`, `agent-output`, `issue`, `best-practice`, `client-customization`, `pattern`, `workflow`, `general`

**Best practice categories**: `backend`, `frontend`, `umbraco`, `nextjs`, `accessibility`, `performance`, `security`, `testing`, `general`

**Issue severity**: `low`, `medium`, `high`, `critical`

**Validation status**: `pending`, `passed`, `failed`, `needs-review`

**Shared vs Custom**: `shared` (in seed project), `custom` (client-specific)

## Tips

- **Search first** before implementing - someone may have solved it already
- **Use tags** for easier searching: `tags: "frontend,gsap,animation"`
- **Record WHY** not just WHAT - rationale matters for future context
- **Document failures** - prevents repeating mistakes
- **Update statuses** - mark issues as resolved when fixed
- **Track client work** - critical for seed project management

## Full Docs

See `README.md` in this directory for complete documentation.
