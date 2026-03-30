package verify

import "context"

// CheckStatus is the result of running a single check.
type CheckStatus string

const (
	CheckStatusPassed  CheckStatus = "passed"
	CheckStatusFailed  CheckStatus = "failed"
	CheckStatusSkipped CheckStatus = "skipped"
	CheckStatusWarning CheckStatus = "warning"
)

// Check is a single health check that can be run during verify phase.
type Check struct {
	ID          string
	Description string
	Run         func(context.Context) error
	// Soft marks this check as non-blocking: errors become warnings instead of failures.
	Soft bool
}

// CheckResult holds the outcome of a single Check.
type CheckResult struct {
	ID          string
	Description string
	Status      CheckStatus
	Error       string
}

// RunChecks executes all checks in order and returns their results.
// A nil Run function results in a Skipped status.
func RunChecks(ctx context.Context, checks []Check) []CheckResult {
	results := make([]CheckResult, 0, len(checks))

	for _, check := range checks {
		result := CheckResult{
			ID:          check.ID,
			Description: check.Description,
		}

		if check.Run == nil {
			result.Status = CheckStatusSkipped
			result.Error = "check not implemented"
			results = append(results, result)
			continue
		}

		if err := check.Run(ctx); err != nil {
			if check.Soft {
				result.Status = CheckStatusWarning
			} else {
				result.Status = CheckStatusFailed
			}
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		result.Status = CheckStatusPassed
		results = append(results, result)
	}

	return results
}

// AnyFailed reports whether any result has Failed status.
func AnyFailed(results []CheckResult) bool {
	for _, r := range results {
		if r.Status == CheckStatusFailed {
			return true
		}
	}
	return false
}
