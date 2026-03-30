---
name: sdd-apply
description: Implement tasks one by one following the design, test each before proceeding
triggers:
  - /sdd-apply
  - /sdd apply
---

# SDD Apply — Implementation

Execute tasks from `.sdd/tasks.md` one at a time. Never batch.

## Process Per Task
1. Read the task definition
2. Write the implementation
3. Write tests (or run existing ones)
4. Verify the done condition is met
5. Move to next task only after current passes

## Strict TDD Mode (if active)
Follow RED → GREEN → REFACTOR:
1. **RED**: Write a failing test that describes the behavior
2. **GREEN**: Write the minimal code to make the test pass
3. **REFACTOR**: Clean up without breaking the test

## Rules
- Never implement task N+1 before task N passes its tests
- Commit after each completed task: `git commit -m "feat: <task name>"`
- If a task is blocked, surface it immediately — do not skip

## Memory
`mem_save("SDD apply: completed task <N> - <name>", topic_key="sdd-current")`
