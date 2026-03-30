---
name: branch-pr
description: Create git branches and pull requests following project conventions
triggers:
  - /branch-pr
  - /pr
---

# Branch & PR Creation

## Branch Creation

1. Read the project's branch convention from memory or ask the user:
   - Base branch (main/master/develop)
   - Prefix format (feat/, fix/, chore/, etc.)
   - Naming pattern (kebab-case, snake_case, etc.)

2. Create the branch:
```bash
git checkout <base-branch>
git pull origin <base-branch>
git checkout -b <prefix>/<description>
```

## PR Creation

Before creating the PR:
1. Review all commits since the base branch: `git log <base>...HEAD --oneline`
2. Review the full diff: `git diff <base>...HEAD`
3. Check for any unintended changes

PR title format: `<type>: <concise description>` (under 70 chars)

PR body must include:
- **Summary**: 1-3 bullet points of what changed and why
- **Test plan**: Checklist of how to verify the changes
- **Breaking changes**: Any API or behavior changes (if applicable)

```bash
gh pr create --title "<title>" --body "<body>"
```

## Rules
- Never force-push to main/master/develop
- Always create PRs from a feature branch, never directly from base
- One logical change per PR — split if needed
- All tests must pass before creating the PR
