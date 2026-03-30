---
name: issue-creation
description: Create well-structured GitHub/Linear issues
triggers:
  - /issue
  - /create-issue
---

# Issue Creation

## Before Creating

1. Check if a similar issue already exists
2. Determine the issue type: Bug / Feature / Tech Debt / Documentation
3. Gather: reproduction steps (bugs), acceptance criteria (features), motivation (all)

## Bug Report Template

```markdown
## Bug Description
[Clear, one-sentence description of the bug]

## Steps to Reproduce
1. [First step]
2. [Second step]
3. [See error]

## Expected Behavior
[What should happen]

## Actual Behavior
[What actually happens]

## Environment
- OS:
- Version/Branch:
- Relevant config:

## Additional Context
[Logs, screenshots, related issues]
```

## Feature Request Template

```markdown
## Summary
[One paragraph: what, why, for whom]

## Motivation
[The problem this solves — not the solution]

## Proposed Solution
[High-level approach]

## Acceptance Criteria
- [ ] [Testable criterion 1]
- [ ] [Testable criterion 2]

## Out of Scope
[What this issue explicitly does NOT include]
```

## Creating the Issue

```bash
gh issue create --title "<title>" --body "<body>" --label "<type>"
```

For Linear: use the project's Linear CLI or web interface with the same structure.
