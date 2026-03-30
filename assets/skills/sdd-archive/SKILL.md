---
name: sdd-archive
description: Archive the SDD workflow — document decisions and clean up
triggers:
  - /sdd-archive
  - /sdd archive
---

# SDD Archive — Knowledge Preservation

Capture what was learned and clean up the workspace.

## Steps
1. **Write the decision record** — Why was this approach chosen over alternatives?
2. **Update documentation** — README, architecture docs, ADRs if applicable
3. **Archive SDD files** — Move `.sdd/` to `.sdd/archive/<date>-<feature>/`
4. **Clean up** — Remove temporary files, unused branches
5. **Final memory save** — Persist the full summary

## Output: Engram memory entry
```
What: Implemented <feature>
Why: <reason>
Where: <key files>
Learned: <key insights for future sessions>
```

## Memory
`mem_save("ARCHIVED: <feature> - <one sentence summary>", topic_key="sdd-archive-<date>")`
`mem_save("" , topic_key="sdd-current")  // clear current session context`
