---
name: sdd-explore
description: Explore the codebase to understand context before proposing solutions
triggers:
  - /sdd-explore
  - /sdd explore
---

# SDD Explore — Codebase Analysis

Map the existing system before designing anything new.

## Steps
1. **Read the relevant code** — Find all files related to the feature area.
2. **Map dependencies** — What does this code depend on? What depends on it?
3. **Identify patterns** — What conventions does the codebase use?
4. **Find extension points** — Where does the new feature fit?
5. **Flag blockers** — What existing code might conflict or need refactoring?

## Output
Create `.sdd/explore.md` with:
- Relevant files and their responsibilities
- Dependency map
- Existing patterns to follow
- Extension points identified
- Potential conflicts

## Memory
`mem_save("SDD explore: <key findings>", topic_key="sdd-current")`
