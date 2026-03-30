---
name: sdd-init
description: Initialize a new SDD workflow — capture requirements and define success criteria
triggers:
  - /sdd-init
  - /sdd new
---

# SDD Init — Initialize Workflow

Capture the full context of what needs to be built before writing a single line of code.

## Steps
1. **Clarify the goal** — Ask clarifying questions until the requirement is unambiguous.
2. **Define success criteria** — What does "done" look like? List testable acceptance criteria.
3. **Identify constraints** — Performance, security, compatibility, timeline.
4. **List risks** — What could go wrong? What unknowns exist?
5. **Scope the work** — What is explicitly OUT of scope?

## Output
Create `.sdd/init.md` with:
- Goal statement (one paragraph)
- Success criteria (bulleted, testable)
- Constraints
- Risks and unknowns
- Out of scope

## Memory
Save to Engram: `mem_save("SDD init completed for: <goal>", topic_key="sdd-current")`
