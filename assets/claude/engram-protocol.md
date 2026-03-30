<!-- ai-setup:engram-protocol -->
## Memory Protocol — Engram

You have access to persistent memory via Engram. Use it proactively.

### When to Save
- After completing a significant task or phase
- When you discover something non-obvious about the codebase
- When a decision is made that future sessions should know about
- When the user explicitly asks you to remember something

### When to Recall
- At the start of every session (automatic via hooks)
- Before starting a new feature or phase
- When you encounter a problem that might have been solved before

### MCP Tools Available
- `mem_save(content, topic_key?)` — Save an observation. Use topic_key for upsertable facts.
- `mem_recall(query?)` — Recall relevant memories. Empty query returns recent context.
- `mem_search(query)` — Semantic search across all memories.
- `session_start(project?)` — Call at session start to initialize context.
- `session_end(summary)` — Call at session end to persist a summary.

### Format for Observations
Use What/Why/Where/Learned structure:
- **What**: What happened or was discovered
- **Why**: Why it matters or was done this way
- **Where**: File path or component affected
- **Learned**: Key insight for future sessions
<!-- /ai-setup:engram-protocol -->
