---
name: sdd-propose
description: Generate and evaluate 2-3 solution approaches with trade-off analysis
triggers:
  - /sdd-propose
  - /sdd propose
---

# SDD Propose — Solution Approaches

Generate multiple approaches and evaluate them honestly before picking one.

## Steps
1. Generate **2-3 distinct approaches** — not variations of the same idea.
2. For each approach, evaluate:
   - Pros and cons
   - Implementation complexity
   - Performance implications
   - Maintenance cost
   - Risk level
3. **Recommend one approach** with clear justification.
4. Get explicit approval before proceeding to spec.

## Output Format
For each approach:
```
### Approach N: <Name>
**Description**: ...
**Pros**: ...
**Cons**: ...
**Complexity**: Low/Medium/High
**Recommended**: Yes/No — because...
```

## Memory
`mem_save("SDD proposal: chose <approach> because <reason>", topic_key="sdd-current")`
