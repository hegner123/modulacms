# Team Memory MCP Server

A persistent team knowledge base using SQLite and the Model Context Protocol (MCP). Stores design decisions, component contracts, best practices, and shared team knowledge.

## Purpose

This MCP server provides a **shared team memory** that persists across sessions and team members. It's designed to:

- **Capture institutional knowledge** - Design decisions, architectural patterns, and "why" behind choices
- **Track agent outputs** - Component contracts, generated code, and validation results
- **Share best practices** - Learnings discovered during development
- **Document issues** - Known problems and workarounds to avoid duplicate debugging
- **Manage client customizations** - Track what's seed-shared vs. client-specific (critical for this seed project)
- **Onboard new team members** - Contractors can quickly understand context and decisions

## Why This Matters for This Project

Given the project context:
- **4 senior developers + 3-7 rotating contractors** - Knowledge needs to persist across team changes
- **Contractors rotate every ~6 months** - Institutional memory would otherwise be lost
- **Seed project → 20+ client deployments** - Need to track what's shared vs. customized
- **AI agent workflow** - Need to record what worked, what didn't, and validation results
- **Architecture evolving rapidly** - Must document decisions and rationale as they're made

## Installation

### 1. Install Dependencies

```bash
cd ai/mcp
npm install
```

### 2. Build the Server

```bash
npm run build
```

This compiles the TypeScript source to JavaScript in the `dist/` directory.

### 3. Enable in Claude Code

The server is already configured in `.mcp.json` at the project root. When you start Claude Code in this project, you'll be prompted to enable the `team-memory` server. Approve it to start using the team memory tools.

Alternatively, you can enable it manually in `.claude/settings.local.json`:

```json
{
  "enableAllProjectMcpServers": true
}
```

Or enable just this server:

```json
{
  "enabledMcpjsonServers": ["team-memory"]
}
```

### 4. Verify It's Working

Once enabled, you should see the team memory tools available in Claude Code. You can test by asking:

```
Add a memory: "Team memory system installed and working"
```

## Database Location

The SQLite database is stored at:
```
ai/mcp/team-memory.db
```

**IMPORTANT**: This database file is checked into the repository so the whole team shares the same memory. Do not add it to `.gitignore`.

## Available Tools

### Core Memory Management

#### `add_memory`
Add a general memory/note to the knowledge base.

**When to use**: Recording any team-relevant information that doesn't fit other specialized categories.

**Parameters**:
- `title` (required): Short, descriptive title
- `content` (required): Detailed content
- `category` (required): One of: `design-decision`, `architecture`, `component`, `agent-output`, `issue`, `best-practice`, `client-customization`, `pattern`, `workflow`, `general`
- `tags` (optional): Comma-separated tags for categorization
- `author` (required): Who is adding this (name or "AI Agent")
- `metadata` (optional): JSON string with additional structured data

**Example**:
```
Add a memory with title "GSAP animation pattern" and content "Always use gsap.context() in React components with cleanup in useLayoutEffect" and category "best-practice" and author "AI Agent" and tags "frontend,animation,gsap"
```

#### `search_memories`
Search team memories using full-text search or filters.

**Parameters**:
- `query` (optional): Search query (searches title, content, and tags)
- `category` (optional): Filter by category
- `tags` (optional): Filter by tags (comma-separated)
- `limit` (optional): Maximum number of results (default: 10)

**Example**:
```
Search memories for "GSAP animation"
```

### Design Decisions

#### `record_design_decision`
Record an architectural or design decision with rationale.

**When to use**: Whenever the team makes a significant technical decision.

**Parameters**:
- `decision` (required): The decision that was made
- `rationale` (required): Why this decision was made
- `alternatives` (optional): What alternatives were considered
- `impact` (optional): What this decision affects
- `made_by` (required): Who made this decision
- `tags` (optional): Comma-separated tags

**Example**:
```
Record design decision: "Use GSAP for all animations instead of CSS" with rationale "Better cross-browser consistency, runtime control, and timeline sequencing for complex interactions" and alternatives "CSS animations, Framer Motion" and impact "All frontend components, animation guidelines" and made_by "Senior Lead" and tags "frontend,animation,architecture"
```

### Component Contracts

#### `store_component_contract`
Store a component contract from agent workflow.

**When to use**: After vision agent analyzes a design or backend agent generates a contract.

**Parameters**:
- `component_name` (required): Name of the component
- `contract_json` (required): Full component contract as JSON string
- `design_image_path` (optional): Path to design image
- `generated_files` (optional): JSON array of generated file paths
- `validation_status` (required): One of: `pending`, `passed`, `failed`, `needs-review`
- `validation_notes` (optional): Notes about validation
- `agent_version` (optional): Agent version that generated this
- `tags` (optional): Comma-separated tags

**Example**:
```
Store component contract for "HeroSection" with contract_json "{...}" and validation_status "passed" and validation_notes "All files generated successfully, screenshots match design"
```

### Agent Execution Tracking

#### `log_agent_execution`
Log an agent execution with inputs, outputs, and success status.

**When to use**: After running any agent in the workflow.

**Parameters**:
- `agent_name` (required): Name of the agent that executed
- `task_description` (required): What task the agent was performing
- `input_summary` (optional): Summary of inputs
- `output_summary` (optional): Summary of outputs
- `success` (required): Whether execution succeeded (true/false)
- `error_message` (optional): Error message if failed
- `execution_time_ms` (optional): Execution time in milliseconds
- `artifacts` (optional): JSON array of generated artifacts
- `executed_by` (required): Who ran the agent
- `tags` (optional): Comma-separated tags

**Example**:
```
Log agent execution: agent_name "backend-agent" with task_description "Generate UDA files for HeroSection" and success true and executed_by "AI Agent" and artifacts '["data-type__hero-section.uda"]'
```

### Known Issues

#### `record_known_issue`
Record a known issue with optional workaround.

**When to use**: Documenting bugs, limitations, or gotchas the team should know about.

**Parameters**:
- `title` (required): Short description of the issue
- `description` (required): Detailed description
- `workaround` (optional): Workaround if available
- `severity` (optional): One of: `low`, `medium`, `high`, `critical` (default: medium)
- `affected_components` (optional): Comma-separated list of affected components
- `reported_by` (required): Who reported this issue
- `tags` (optional): Comma-separated tags

**Example**:
```
Record known issue: title "Next.js Image with Umbraco media paths" and description "next/image doesn't work with dynamic Umbraco media paths without custom loader" and workaround "Use custom loader in next.config.js or use standard img tag for CMS images" and severity "medium" and reported_by "Senior Dev"
```

#### `list_open_issues`
List all open or investigating known issues, optionally filtered by severity.

**Parameters**:
- `severity` (optional): Filter by severity

**Example**:
```
List open issues with severity "high"
```

### Best Practices

#### `add_best_practice`
Record a best practice discovered during development.

**When to use**: Sharing team learnings, patterns that work well, anti-patterns to avoid.

**Parameters**:
- `title` (required): Short title for the best practice
- `description` (required): Detailed description of the practice
- `example_code` (optional): Code snippet demonstrating the practice
- `context` (optional): When/where to apply this
- `category` (required): One of: `backend`, `frontend`, `umbraco`, `nextjs`, `accessibility`, `performance`, `security`, `testing`, `general`
- `learned_from` (optional): What situation taught us this
- `added_by` (required): Who is adding this
- `tags` (optional): Comma-separated tags

**Example**:
```
Add best practice: title "Always clean up GSAP ScrollTrigger" and description "In React components, always kill ScrollTrigger instances in useLayoutEffect cleanup to prevent memory leaks" and example_code "useLayoutEffect(() => { const ctx = gsap.context(() => {...}); return () => ctx.revert(); }, []);" and category "frontend" and added_by "AI Agent"
```

### Client Customizations

#### `record_client_customization`
Track client-specific customizations.

**When to use**: Critical for seed project management - track what should stay in seed vs. what's client-specific.

**Parameters**:
- `client_name` (required): Name of the client
- `customization_type` (required): One of: `component`, `theme`, `feature`, `workflow`, `integration`, `configuration`
- `description` (required): What was customized
- `implementation_notes` (optional): How it was implemented
- `files_affected` (optional): JSON array of affected file paths
- `shared_vs_custom` (required): One of: `shared` (should be in seed), `custom` (client-specific)
- `tags` (optional): Comma-separated tags

**Example**:
```
Record client customization for client_name "MarineClient1" with customization_type "theme" and description "Custom brand colors: Navy blue (#003366) and seafoam green (#20B2AA)" and shared_vs_custom "custom" and files_affected '["styles/theme.scss", "components/Header.tsx"]'
```

### Architectural Patterns

#### `add_pattern`
Document an architectural or code pattern and when to use it.

**When to use**: Documenting reusable patterns, design patterns, or conventions.

**Parameters**:
- `pattern_name` (required): Name of the pattern
- `description` (required): What the pattern is
- `use_cases` (optional): When to use this pattern
- `example_implementation` (optional): Code example or reference
- `pros` (optional): Advantages of this pattern
- `cons` (optional): Disadvantages or tradeoffs
- `related_patterns` (optional): Comma-separated related patterns
- `category` (required): One of: `backend`, `frontend`, `fullstack`, `data-access`, `api`, `ui`, `state-management`
- `tags` (optional): Comma-separated tags

**Example**:
```
Add pattern: pattern_name "ApiSafeConverter Pattern" and description "Transform Umbraco content nodes to JSON-safe models for frontend consumption" and use_cases "Every document type that needs to be exposed via API" and category "backend" and tags "umbraco,api,converter"
```

### Activity and History

#### `get_recent_activity`
Get recent team activity across all memory types.

**When to use**: Onboarding new team members, catching up after time away.

**Parameters**:
- `days` (optional): Number of days to look back (default: 7)
- `limit` (optional): Maximum items to return (default: 20)

**Example**:
```
Get recent activity from the last 14 days
```

#### `get_component_history`
Get the full history and context for a specific component.

**When to use**: Understanding a component's evolution, contracts, and related decisions.

**Parameters**:
- `component_name` (required): Name of the component to query

**Example**:
```
Get component history for "HeroSection"
```

## Usage Workflows

### Agent Workflow: Design → Backend → Frontend

1. **Vision agent analyzes design**:
   ```
   Store component contract for "FeatureGrid" with contract_json "{...}" and validation_status "pending"
   ```

2. **Backend agent generates UDA files**:
   ```
   Log agent execution: agent_name "backend-agent" with task_description "Generate UDA files for FeatureGrid" and success true
   ```

3. **Frontend agent generates components**:
   ```
   Log agent execution: agent_name "frontend-agent" with task_description "Generate React component for FeatureGrid" and success true
   ```

4. **Update contract validation**:
   ```
   Store component contract for "FeatureGrid" with validation_status "passed" and validation_notes "Frontend matches design, responsive on all breakpoints"
   ```

### Recording Decisions

When making architectural decisions:

```
Record design decision: "Use Server Components by default in Next.js" with rationale "Reduces bundle size, improves performance, follows Next.js 15 best practices" and alternatives "Client Components everywhere, Mixed approach" and impact "All new page components" and made_by "Senior Lead"
```

### Documenting Issues

When encountering a bug or limitation:

```
Record known issue: title "Umbraco Forms not accessible in Server Components" and description "Umbraco Forms API requires client-side state, cannot be used in Next.js Server Components" and workaround "Create Client Component wrapper for Umbraco Forms" and severity "medium" and reported_by "Contractor"
```

### Sharing Best Practices

When learning something useful:

```
Add best practice: title "Lazy load images below fold" and description "Use loading='lazy' for images below the fold to improve initial page load performance" and category "performance" and added_by "Senior Dev"
```

### Client Customization Tracking

When implementing client-specific features:

```
Record client customization for client_name "RVClient1" with customization_type "feature" and description "Custom RV model comparison tool with side-by-side specs" and shared_vs_custom "custom" and implementation_notes "Uses custom API endpoint /api/rv-compare"
```

## Maintenance

### Rebuilding After Changes

If you modify the server code, rebuild it:

```bash
cd ai/mcp
npm run build
```

Then restart Claude Code to reload the server.

### Backing Up the Database

The database is checked into git, but you can also manually back it up:

```bash
cp ai/mcp/team-memory.db ai/mcp/team-memory.db.backup
```

### Querying Directly (Advanced)

You can query the database directly using SQLite tools:

```bash
sqlite3 ai/mcp/team-memory.db
```

Example queries:

```sql
-- See all design decisions
SELECT * FROM design_decisions ORDER BY date_decided DESC;

-- Find high-severity open issues
SELECT * FROM known_issues WHERE severity = 'critical' AND status = 'open';

-- Search memories
SELECT * FROM memories WHERE content LIKE '%GSAP%';
```

## Schema Overview

The database includes these tables:

- **memories** - General knowledge and notes with full-text search
- **design_decisions** - Architectural decisions with rationale and alternatives
- **component_contracts** - Component contracts from agent workflows
- **agent_executions** - Agent execution history and results
- **known_issues** - Known bugs and workarounds with severity tracking
- **best_practices** - Team learnings and coding patterns
- **client_customizations** - Client-specific changes (seed project critical)
- **patterns** - Architectural and code patterns

See `schema.sql` for full schema details.

## Tips for Effective Use

1. **Record decisions when they're made** - Don't wait, capture context while it's fresh
2. **Use tags liberally** - Makes searching easier later
3. **Include rationale** - Future you (or contractors) will thank you
4. **Document what didn't work** - Prevents trying the same failed approach twice
5. **Update issue statuses** - Mark issues as resolved when fixed
6. **Track client customizations** - Critical for maintaining the seed project
7. **Search before implementing** - Check if someone already solved your problem
8. **Use descriptive titles** - Makes search results more useful

## Team Adoption

### For Senior Developers
- Record architectural decisions and rationale
- Document patterns and conventions
- Track what works at scale
- Review agent outputs and record results

### For Contractors
- Search before asking questions
- Add best practices you discover
- Record issues and workarounds
- Document client customizations

### For AI Agents
- Store component contracts after generation
- Log execution results
- Record validation outcomes
- Search for existing patterns before generating new ones

## Troubleshooting

### Server not starting
1. Check that dependencies are installed: `npm install` in `ai/mcp/`
2. Verify the build succeeded: `npm run build`
3. Check for errors in Claude Code's MCP server logs

### Database locked errors
SQLite uses WAL mode for better concurrent access. If you see lock errors, ensure no other process is accessing the database.

### Search not working
The full-text search index should be automatically maintained. If search isn't working, the schema may not have been applied correctly. Delete `team-memory.db` and restart the server to recreate it.

## Future Enhancements

Potential additions:
- Web UI for browsing team memory
- Export to markdown documentation
- Integration with Slack/Teams for notifications
- Automatic memory extraction from git commits
- Client-specific memory views (filter by client)
- Memory metrics and analytics

---

**Questions?** Check the troubleshooting section or ask in team chat.

**Contributing?** Suggest improvements or additional memory categories to the team lead.
