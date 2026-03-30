---
name: sdd-tasks
description: Break the design into atomic, independently testable implementation tasks
triggers:
  - /sdd-tasks
  - /sdd tasks
---

# SDD Tasks — Task Breakdown

Break the design into the smallest independently deliverable units.

## Rules
- Each task must be completable in one focused session
- Each task must have a clear, testable done condition
- Tasks must be ordered by dependency (no task should require an incomplete predecessor)
- No task should modify more than 3 files

## Output: `.sdd/tasks.md`
Each task entry:
```
### Task N: <Name>
**What**: [one sentence]
**Files to create/modify**: [list]
**Done when**: [testable criterion]
**Depends on**: Task M (or "none")
```

## Memory
`mem_save("SDD tasks defined: <N> tasks for <feature>", topic_key="sdd-current")`
