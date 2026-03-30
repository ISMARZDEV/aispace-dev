---
name: sdd-verify
description: Run full test suite and validate implementation against the specification
triggers:
  - /sdd-verify
  - /sdd verify
---

# SDD Verify — Validation

Verify the implementation is complete, correct, and matches the spec.

## Checklist
- [ ] All tasks from `.sdd/tasks.md` are marked complete
- [ ] Full test suite passes with no failures
- [ ] All acceptance criteria from `.sdd/init.md` are met
- [ ] No regressions in existing tests
- [ ] Code coverage meets project standards
- [ ] Error paths are tested, not just happy paths
- [ ] Performance is acceptable (no obvious bottlenecks)

## If Verification Fails
- Document what failed and why
- Create a fix task and re-enter the apply phase
- Do not proceed to archive until all criteria pass

## Memory
`mem_save("SDD verify: PASSED for <feature>", topic_key="sdd-current")`
