# Policy Evaluation Algorithm

## Current Algorithm: Deny-Override

The ABAC system uses a **Deny-Override** algorithm for policy evaluation. This is a security-focused approach where any explicit Deny decision takes precedence over Allow decisions.

### Algorithm Flow

```
┌─────────────────────────────────────────┐
│  1. Load All Enabled Policies           │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│  2. Evaluate Each Policy Statement      │
│     - Check Action Match                │
│     - Check Resource Match              │
│     - Check Conditions                  │
└─────────────────┬───────────────────────┘
                  │
         ┌────────┴────────┐
         │                 │
    ┌────▼────┐      ┌────▼────┐
    │  Deny   │      │  Allow  │
    │  Match? │      │  Match? │
    └────┬────┘      └────┬────┘
         │                │
         │ YES            │ YES
         │                │
    ┌────▼────────────────▼────┐
    │  Deny Override           │
    │  ↓                       │
    │  Return DENY Immediately │
    └──────────────────────────┘
                  │
                  │ No Deny
                  │
         ┌────────▼────────┐
         │  Any Allow?     │
         │  YES → PERMIT   │
         │  NO  → DENY     │
         └─────────────────┘
```

## Implementation

**File**: `evaluator/core/pdp.go`

```go
func (pdp *PolicyDecisionPoint) evaluateNewPolicies(policies []*models.Policy, context map[string]interface{}) *models.Decision {
    // Step 1: Collect all matching statements
    for _, policy := range policies {
        if !policy.Enabled {
            continue
        }

        for _, statement := range policy.Statement {
            if pdp.evaluateStatement(statement, context) {
                // Step 2: Apply Deny-Override
                if statement.Effect == "Deny" {
                    return &models.Decision{
                        Result: "deny",
                        Reason: "Denied by statement: " + statement.Sid,
                    }
                }
            }
        }
    }

    // Step 3: If we have any Allow statements, return allow
    if len(matchedStatements) > 0 {
        return &models.Decision{
            Result: "permit",
        }
    }

    // Step 4: Default deny (implicit deny)
    return &models.Decision{
        Result: "deny",
        Reason: "No matching policies found (implicit deny)",
    }
}
```

## Key Characteristics

### 1. **Deny Always Wins**
- If ANY policy statement evaluates to Deny, the final result is Deny
- This happens immediately - no further evaluation needed
- Most secure approach: "deny by default, allow explicitly"

### 2. **Simple Evaluation Order**
- Policies are evaluated in database order
- Order doesn't matter because Deny-Override guarantees consistency
- No priority field needed for Deny-Override algorithm

### 3. **Implicit Deny**
- If no policies match → Deny
- If only Deny policies match → Deny  
- If at least one Allow and no Deny → Permit

## Example Scenarios

### Scenario 1: Conflicting Policies

```json
{
  "policies": [
    {
      "id": "allow-read",
      "effect": "Allow",
      "action": "document:read",
      "resource": "/documents/*"
    },
    {
      "id": "deny-confidential",
      "effect": "Deny",
      "action": "document:read",
      "resource": "/documents/confidential/*"
    }
  ]
}
```

**Request**: Read `/documents/confidential/salary.pdf`

**Result**: **DENY** ❌

**Reason**: Both policies match, but Deny-Override means the Deny policy wins.

### Scenario 2: Multiple Allow Policies

```json
{
  "policies": [
    {
      "id": "allow-engineering",
      "effect": "Allow",
      "action": "document:read",
      "condition": { "department": "Engineering" }
    },
    {
      "id": "allow-managers",
      "effect": "Allow",
      "action": "document:read",
      "condition": { "role": "Manager" }
    }
  ]
}
```

**Request**: Engineering employee reads document

**Result**: **PERMIT** ✅

**Reason**: At least one Allow matches, no Deny policies.

### Scenario 3: No Matching Policies

```json
{
  "policies": [
    {
      "id": "allow-engineering",
      "effect": "Allow",
      "action": "document:read",
      "condition": { "department": "Engineering" }
    }
  ]
}
```

**Request**: Finance employee reads document

**Result**: **DENY** ❌

**Reason**: No policies match → Implicit Deny.

## Alternative Algorithms (Not Implemented)

### First-Match
- Stop at first matching policy
- Order matters (would need priority field)
- Less secure (easy to misconfigure)

### Permit-Override
- Any Allow wins over Deny
- Opposite of current approach
- Less secure for most use cases

### Priority-Based
- Sort policies by priority field
- Evaluate in order
- First match wins
- Requires explicit priority management

## Why Deny-Override?

✅ **Security First**: Default deny, explicit allow  
✅ **Simple**: No complex ordering logic  
✅ **Predictable**: Same result regardless of policy order  
✅ **Fast**: Can short-circuit on first Deny  
✅ **Industry Standard**: Used by AWS IAM, Azure RBAC, etc.

## Performance Characteristics

- **Best Case**: O(1) - First policy is a Deny match
- **Average Case**: O(n) - Need to check all policies for Deny
- **Worst Case**: O(n) - Check all policies, all are Allow or no match

Where n = number of enabled policies

## Future Enhancements

If you need priority-based evaluation in the future:

1. Add a `priority` field to `models.Policy`
2. Sort policies before evaluation by priority
3. Implement First-Match or Priority-Override algorithm
4. Update database schema to include priority column
4. Update documentation

## Testing

See test files:
- `evaluator/core/pdp_test.go`
- `evaluator/core/integration_test.go`

Key test: **TestImprovedPDP_RealWorldScenarios** demonstrates Deny-Override in action.

## Related Documentation

- [Condition Guide](CONDITION_GUIDE.md)
- [Policy Format](../README.md#policy-format)
- [AWS IAM Policy Evaluation Logic](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_evaluation-logic.html) (similar approach)

