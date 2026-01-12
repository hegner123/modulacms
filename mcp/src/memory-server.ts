#!/usr/bin/env node

/**
 * Team Memory MCP Server
 *
 * Provides persistent team memory storage using SQLite for the Umbraco + Next.js project.
 * Stores design decisions, component contracts, best practices, and team knowledge.
 */

import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
  Tool,
} from '@modelcontextprotocol/sdk/types.js';
import Database from 'better-sqlite3';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';
import { readFileSync } from 'fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const DB_PATH = join(__dirname, '..', 'team-memory.db');

// Initialize database
const db = new Database(DB_PATH);
db.pragma('journal_mode = WAL');

// Load and execute schema
const schemaPath = join(__dirname, '..', 'schema.sql');
const schema = readFileSync(schemaPath, 'utf-8');
db.exec(schema);

const server = new Server(
  {
    name: 'team-memory',
    version: '1.0.0',
  },
  {
    capabilities: {
      tools: {},
    },
  }
);

// Define available tools
const tools: Tool[] = [
  {
    name: 'add_memory',
    description: 'Add a new memory/note to the team knowledge base. Use for general knowledge, decisions, patterns, or any team-relevant information.',
    inputSchema: {
      type: 'object',
      properties: {
        title: { type: 'string', description: 'Short, descriptive title for the memory' },
        content: { type: 'string', description: 'Detailed content of the memory' },
        category: {
          type: 'string',
          enum: ['design-decision', 'architecture', 'component', 'agent-output', 'issue', 'best-practice', 'client-customization', 'pattern', 'workflow', 'general'],
          description: 'Category of the memory'
        },
        tags: { type: 'string', description: 'Comma-separated tags for categorization (optional)' },
        author: { type: 'string', description: 'Who is adding this memory (name or "AI Agent")' },
        metadata: { type: 'string', description: 'JSON string with additional structured data (optional)' }
      },
      required: ['title', 'content', 'category', 'author']
    }
  },
  {
    name: 'search_memories',
    description: 'Search team memories using full-text search or filters. Returns relevant memories matching the query.',
    inputSchema: {
      type: 'object',
      properties: {
        query: { type: 'string', description: 'Search query (searches title, content, and tags)' },
        category: { type: 'string', description: 'Filter by category (optional)' },
        tags: { type: 'string', description: 'Filter by tags (comma-separated, optional)' },
        limit: { type: 'number', description: 'Maximum number of results (default: 10)' }
      },
      required: []
    }
  },
  {
    name: 'record_design_decision',
    description: 'Record an architectural or design decision with rationale. Important for tracking why choices were made.',
    inputSchema: {
      type: 'object',
      properties: {
        decision: { type: 'string', description: 'The decision that was made' },
        rationale: { type: 'string', description: 'Why this decision was made' },
        alternatives: { type: 'string', description: 'What alternatives were considered (optional)' },
        impact: { type: 'string', description: 'What this decision affects (optional)' },
        made_by: { type: 'string', description: 'Who made this decision' },
        tags: { type: 'string', description: 'Comma-separated tags (optional)' }
      },
      required: ['decision', 'rationale', 'made_by']
    }
  },
  {
    name: 'store_component_contract',
    description: 'Store a component contract from agent workflow. Links design specs to generated code.',
    inputSchema: {
      type: 'object',
      properties: {
        component_name: { type: 'string', description: 'Name of the component' },
        contract_json: { type: 'string', description: 'Full component contract as JSON string' },
        design_image_path: { type: 'string', description: 'Path to design image (optional)' },
        generated_files: { type: 'string', description: 'JSON array of generated file paths (optional)' },
        validation_status: {
          type: 'string',
          enum: ['pending', 'passed', 'failed', 'needs-review'],
          description: 'Validation status'
        },
        validation_notes: { type: 'string', description: 'Notes about validation (optional)' },
        agent_version: { type: 'string', description: 'Agent version that generated this (optional)' },
        tags: { type: 'string', description: 'Comma-separated tags (optional)' }
      },
      required: ['component_name', 'contract_json', 'validation_status']
    }
  },
  {
    name: 'log_agent_execution',
    description: 'Log an agent execution with inputs, outputs, and success status. Useful for tracking what works.',
    inputSchema: {
      type: 'object',
      properties: {
        agent_name: { type: 'string', description: 'Name of the agent that executed' },
        task_description: { type: 'string', description: 'What task the agent was performing' },
        input_summary: { type: 'string', description: 'Summary of inputs (optional)' },
        output_summary: { type: 'string', description: 'Summary of outputs (optional)' },
        success: { type: 'boolean', description: 'Whether execution succeeded' },
        error_message: { type: 'string', description: 'Error message if failed (optional)' },
        execution_time_ms: { type: 'number', description: 'Execution time in milliseconds (optional)' },
        artifacts: { type: 'string', description: 'JSON array of generated artifacts (optional)' },
        executed_by: { type: 'string', description: 'Who ran the agent' },
        tags: { type: 'string', description: 'Comma-separated tags (optional)' }
      },
      required: ['agent_name', 'task_description', 'success', 'executed_by']
    }
  },
  {
    name: 'record_known_issue',
    description: 'Record a known issue with optional workaround. Helps team avoid duplicate debugging.',
    inputSchema: {
      type: 'object',
      properties: {
        title: { type: 'string', description: 'Short description of the issue' },
        description: { type: 'string', description: 'Detailed description of the issue' },
        workaround: { type: 'string', description: 'Workaround if available (optional)' },
        severity: {
          type: 'string',
          enum: ['low', 'medium', 'high', 'critical'],
          description: 'Issue severity (default: medium)'
        },
        affected_components: { type: 'string', description: 'Comma-separated list of affected components (optional)' },
        reported_by: { type: 'string', description: 'Who reported this issue' },
        tags: { type: 'string', description: 'Comma-separated tags (optional)' }
      },
      required: ['title', 'description', 'reported_by']
    }
  },
  {
    name: 'add_best_practice',
    description: 'Record a best practice discovered during development. Share team learnings.',
    inputSchema: {
      type: 'object',
      properties: {
        title: { type: 'string', description: 'Short title for the best practice' },
        description: { type: 'string', description: 'Detailed description of the practice' },
        example_code: { type: 'string', description: 'Code snippet demonstrating the practice (optional)' },
        context: { type: 'string', description: 'When/where to apply this (optional)' },
        category: {
          type: 'string',
          enum: ['backend', 'frontend', 'umbraco', 'nextjs', 'accessibility', 'performance', 'security', 'testing', 'general'],
          description: 'Category of the best practice'
        },
        learned_from: { type: 'string', description: 'What situation taught us this (optional)' },
        added_by: { type: 'string', description: 'Who is adding this' },
        tags: { type: 'string', description: 'Comma-separated tags (optional)' }
      },
      required: ['title', 'description', 'category', 'added_by']
    }
  },
  {
    name: 'record_client_customization',
    description: 'Track client-specific customizations. Critical for seed project management.',
    inputSchema: {
      type: 'object',
      properties: {
        client_name: { type: 'string', description: 'Name of the client' },
        customization_type: {
          type: 'string',
          enum: ['component', 'theme', 'feature', 'workflow', 'integration', 'configuration'],
          description: 'Type of customization'
        },
        description: { type: 'string', description: 'What was customized' },
        implementation_notes: { type: 'string', description: 'How it was implemented (optional)' },
        files_affected: { type: 'string', description: 'JSON array of affected file paths (optional)' },
        shared_vs_custom: {
          type: 'string',
          enum: ['shared', 'custom'],
          description: 'Should this be in seed (shared) or client-specific (custom)?'
        },
        tags: { type: 'string', description: 'Comma-separated tags (optional)' }
      },
      required: ['client_name', 'customization_type', 'description', 'shared_vs_custom']
    }
  },
  {
    name: 'add_pattern',
    description: 'Document an architectural or code pattern and when to use it.',
    inputSchema: {
      type: 'object',
      properties: {
        pattern_name: { type: 'string', description: 'Name of the pattern' },
        description: { type: 'string', description: 'What the pattern is' },
        use_cases: { type: 'string', description: 'When to use this pattern (optional)' },
        example_implementation: { type: 'string', description: 'Code example or reference (optional)' },
        pros: { type: 'string', description: 'Advantages of this pattern (optional)' },
        cons: { type: 'string', description: 'Disadvantages or tradeoffs (optional)' },
        related_patterns: { type: 'string', description: 'Comma-separated related patterns (optional)' },
        category: {
          type: 'string',
          enum: ['backend', 'frontend', 'fullstack', 'data-access', 'api', 'ui', 'state-management'],
          description: 'Pattern category'
        },
        tags: { type: 'string', description: 'Comma-separated tags (optional)' }
      },
      required: ['pattern_name', 'description', 'category']
    }
  },
  {
    name: 'get_recent_activity',
    description: 'Get recent team activity across all memory types. Useful for onboarding or catching up.',
    inputSchema: {
      type: 'object',
      properties: {
        days: { type: 'number', description: 'Number of days to look back (default: 7)' },
        limit: { type: 'number', description: 'Maximum items to return (default: 20)' }
      },
      required: []
    }
  },
  {
    name: 'get_component_history',
    description: 'Get the full history and context for a specific component.',
    inputSchema: {
      type: 'object',
      properties: {
        component_name: { type: 'string', description: 'Name of the component to query' }
      },
      required: ['component_name']
    }
  },
  {
    name: 'list_open_issues',
    description: 'List all open or investigating known issues, optionally filtered by severity.',
    inputSchema: {
      type: 'object',
      properties: {
        severity: {
          type: 'string',
          enum: ['low', 'medium', 'high', 'critical'],
          description: 'Filter by severity (optional)'
        }
      },
      required: []
    }
  }
];

// Tool handlers
server.setRequestHandler(ListToolsRequestSchema, async () => ({
  tools
}));

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;

  if (!args) {
    throw new Error('Arguments are required for tool calls');
  }

  try {
    switch (name) {
      case 'add_memory': {
        const stmt = db.prepare(`
          INSERT INTO memories (title, content, category, tags, author, metadata)
          VALUES (?, ?, ?, ?, ?, ?)
        `);
        const result = stmt.run(
          args.title,
          args.content,
          args.category,
          args.tags || null,
          args.author,
          args.metadata || null
        );
        return {
          content: [{
            type: 'text',
            text: `Memory added successfully with ID ${result.lastInsertRowid}`
          }]
        };
      }

      case 'search_memories': {
        let query = 'SELECT * FROM memories WHERE 1=1';
        const params: any[] = [];

        if (args.query) {
          query = `
            SELECT m.* FROM memories m
            JOIN memories_fts fts ON m.id = fts.rowid
            WHERE memories_fts MATCH ?
          `;
          params.push(args.query);
        }

        if (args.category) {
          query += ' AND category = ?';
          params.push(args.category);
        }

        if (args.tags) {
          query += ' AND tags LIKE ?';
          params.push(`%${args.tags}%`);
        }

        query += ' ORDER BY created_at DESC LIMIT ?';
        params.push(args.limit || 10);

        const stmt = db.prepare(query);
        const results = stmt.all(...params);

        return {
          content: [{
            type: 'text',
            text: JSON.stringify(results, null, 2)
          }]
        };
      }

      case 'record_design_decision': {
        const stmt = db.prepare(`
          INSERT INTO design_decisions (decision, rationale, alternatives_considered, impact, made_by, tags)
          VALUES (?, ?, ?, ?, ?, ?)
        `);
        const result = stmt.run(
          args.decision,
          args.rationale,
          args.alternatives || null,
          args.impact || null,
          args.made_by,
          args.tags || null
        );
        return {
          content: [{
            type: 'text',
            text: `Design decision recorded with ID ${result.lastInsertRowid}`
          }]
        };
      }

      case 'store_component_contract': {
        const stmt = db.prepare(`
          INSERT INTO component_contracts
          (component_name, contract_json, design_image_path, generated_files, validation_status, validation_notes, agent_version, tags)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `);
        const result = stmt.run(
          args.component_name,
          args.contract_json,
          args.design_image_path || null,
          args.generated_files || null,
          args.validation_status,
          args.validation_notes || null,
          args.agent_version || null,
          args.tags || null
        );
        return {
          content: [{
            type: 'text',
            text: `Component contract stored with ID ${result.lastInsertRowid}`
          }]
        };
      }

      case 'log_agent_execution': {
        const stmt = db.prepare(`
          INSERT INTO agent_executions
          (agent_name, task_description, input_summary, output_summary, success, error_message, execution_time_ms, artifacts, executed_by, tags)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `);
        const result = stmt.run(
          args.agent_name,
          args.task_description,
          args.input_summary || null,
          args.output_summary || null,
          args.success ? 1 : 0,
          args.error_message || null,
          args.execution_time_ms || null,
          args.artifacts || null,
          args.executed_by,
          args.tags || null
        );
        return {
          content: [{
            type: 'text',
            text: `Agent execution logged with ID ${result.lastInsertRowid}`
          }]
        };
      }

      case 'record_known_issue': {
        const stmt = db.prepare(`
          INSERT INTO known_issues
          (title, description, workaround, severity, affected_components, reported_by, tags)
          VALUES (?, ?, ?, ?, ?, ?, ?)
        `);
        const result = stmt.run(
          args.title,
          args.description,
          args.workaround || null,
          args.severity || 'medium',
          args.affected_components || null,
          args.reported_by,
          args.tags || null
        );
        return {
          content: [{
            type: 'text',
            text: `Known issue recorded with ID ${result.lastInsertRowid}`
          }]
        };
      }

      case 'add_best_practice': {
        const stmt = db.prepare(`
          INSERT INTO best_practices
          (title, description, example_code, context, category, learned_from, added_by, tags)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `);
        const result = stmt.run(
          args.title,
          args.description,
          args.example_code || null,
          args.context || null,
          args.category,
          args.learned_from || null,
          args.added_by,
          args.tags || null
        );
        return {
          content: [{
            type: 'text',
            text: `Best practice added with ID ${result.lastInsertRowid}`
          }]
        };
      }

      case 'record_client_customization': {
        const stmt = db.prepare(`
          INSERT INTO client_customizations
          (client_name, customization_type, description, implementation_notes, files_affected, shared_vs_custom, tags)
          VALUES (?, ?, ?, ?, ?, ?, ?)
        `);
        const result = stmt.run(
          args.client_name,
          args.customization_type,
          args.description,
          args.implementation_notes || null,
          args.files_affected || null,
          args.shared_vs_custom,
          args.tags || null
        );
        return {
          content: [{
            type: 'text',
            text: `Client customization recorded with ID ${result.lastInsertRowid}`
          }]
        };
      }

      case 'add_pattern': {
        const stmt = db.prepare(`
          INSERT INTO patterns
          (pattern_name, description, use_cases, example_implementation, pros, cons, related_patterns, category, tags)
          VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `);
        const result = stmt.run(
          args.pattern_name,
          args.description,
          args.use_cases || null,
          args.example_implementation || null,
          args.pros || null,
          args.cons || null,
          args.related_patterns || null,
          args.category,
          args.tags || null
        );
        return {
          content: [{
            type: 'text',
            text: `Pattern added with ID ${result.lastInsertRowid}`
          }]
        };
      }

      case 'get_recent_activity': {
        const days: number = (args.days as number) || 7;
        const limit: number = (args.limit as number) || 20;

        const activities: any[] = [];

        // Get recent memories
        const memoriesStmt = db.prepare(`
          SELECT 'memory' as type, id, title, category, created_at, author as actor
          FROM memories
          WHERE created_at >= datetime('now', '-' || ? || ' days')
          ORDER BY created_at DESC
          LIMIT ?
        `);
        activities.push(...memoriesStmt.all(days, limit));

        // Get recent decisions
        const decisionsStmt = db.prepare(`
          SELECT 'decision' as type, id, decision as title, status as category, date_decided as created_at, made_by as actor
          FROM design_decisions
          WHERE date_decided >= datetime('now', '-' || ? || ' days')
          ORDER BY date_decided DESC
          LIMIT ?
        `);
        activities.push(...decisionsStmt.all(days, limit));

        // Get recent agent executions
        const agentsStmt = db.prepare(`
          SELECT 'agent' as type, id, agent_name || ': ' || task_description as title,
                 CASE WHEN success = 1 THEN 'success' ELSE 'failed' END as category,
                 executed_at as created_at, executed_by as actor
          FROM agent_executions
          WHERE executed_at >= datetime('now', '-' || ? || ' days')
          ORDER BY executed_at DESC
          LIMIT ?
        `);
        activities.push(...agentsStmt.all(days, limit));

        // Sort all activities by date and limit
        activities.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
        const recentActivities = activities.slice(0, limit);

        return {
          content: [{
            type: 'text',
            text: JSON.stringify(recentActivities, null, 2)
          }]
        };
      }

      case 'get_component_history': {
        const componentName = args.component_name;
        const history: any = {
          component_name: componentName,
          contracts: [],
          related_memories: [],
          agent_executions: []
        };

        // Get component contracts
        const contractsStmt = db.prepare(`
          SELECT * FROM component_contracts
          WHERE component_name = ?
          ORDER BY created_at DESC
        `);
        history.contracts = contractsStmt.all(componentName);

        // Get related memories
        const memoriesStmt = db.prepare(`
          SELECT * FROM memories
          WHERE category = 'component' AND (title LIKE ? OR content LIKE ?)
          ORDER BY created_at DESC
        `);
        history.related_memories = memoriesStmt.all(`%${componentName}%`, `%${componentName}%`);

        // Get agent executions
        const agentsStmt = db.prepare(`
          SELECT * FROM agent_executions
          WHERE task_description LIKE ?
          ORDER BY executed_at DESC
        `);
        history.agent_executions = agentsStmt.all(`%${componentName}%`);

        return {
          content: [{
            type: 'text',
            text: JSON.stringify(history, null, 2)
          }]
        };
      }

      case 'list_open_issues': {
        let query = `
          SELECT * FROM known_issues
          WHERE status IN ('open', 'investigating')
        `;
        const params: any[] = [];

        if (args.severity) {
          query += ' AND severity = ?';
          params.push(args.severity);
        }

        query += ' ORDER BY severity DESC, reported_at DESC';

        const stmt = db.prepare(query);
        const issues = stmt.all(...params);

        return {
          content: [{
            type: 'text',
            text: JSON.stringify(issues, null, 2)
          }]
        };
      }

      default:
        return {
          content: [{
            type: 'text',
            text: `Unknown tool: ${name}`
          }],
          isError: true
        };
    }
  } catch (error) {
    return {
      content: [{
        type: 'text',
        text: `Error executing ${name}: ${error instanceof Error ? error.message : String(error)}`
      }],
      isError: true
    };
  }
});

// Start server
async function main() {
  const transport = new StdioServerTransport();
  await server.connect(transport);
  console.error('Team Memory MCP Server running on stdio');
}

main().catch((error) => {
  console.error('Server error:', error);
  process.exit(1);
});
