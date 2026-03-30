---
name: sdd-spec
description: Write a formal technical specification from the approved proposal
triggers:
  - /sdd-spec
  - /sdd spec
---

# SDD Spec — Technical Specification

Turn the approved proposal into a precise, implementable specification.

## Steps
1. Start from the approved approach in `.sdd/propose.md`
2. Define all interfaces, data structures, and contracts
3. Specify error handling and edge cases
4. Define the testing strategy
5. Get explicit approval before proceeding

## Output: `.sdd/spec.md`
- **Overview**: What will be built
- **Interfaces**: Function/API signatures with types
- **Data Models**: Structs, schemas, types
- **Error Cases**: How each error is handled
- **Testing Strategy**: Unit/integration/e2e breakdown
- **Dependencies**: External packages or services required

## Memory
`mem_save("SDD spec approved for: <feature>", topic_key="sdd-current")`
