---
name: sdd-design
description: Design component interfaces, data models, and contracts before implementation
triggers:
  - /sdd-design
  - /sdd design
---

# SDD Design — Component Design

Finalize the detailed design before writing implementation code.

## Steps
1. Design each component's public interface
2. Define data flow between components
3. Identify shared utilities to create or reuse
4. Plan file structure and module organization
5. Review for consistency with existing patterns

## Output: `.sdd/design.md`
- Component diagram (text/ASCII)
- Interface definitions (code stubs)
- Data flow description
- File structure plan
- Reuse opportunities

## Memory
`mem_save("SDD design complete: <key decisions>", topic_key="sdd-current")`
