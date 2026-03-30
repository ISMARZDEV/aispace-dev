package pipeline

import "time"

// Stage identifies which phase of execution a step belongs to.
type Stage string

const (
	StagePrepare  Stage = "prepare"
	StageApply    Stage = "apply"
	StageRollback Stage = "rollback"
)

// StepStatus is the execution state of a single step.
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusRunning    StepStatus = "running"
	StepStatusSucceeded  StepStatus = "succeeded"
	StepStatusFailed     StepStatus = "failed"
	StepStatusRolledBack StepStatus = "rolled_back"
)

// Step is the core unit of work. Components implement this interface.
type Step interface {
	ID() string
	Run() error
}

// RollbackStep is a Step that can undo its own work.
// Only steps that implement this interface are rolled back on failure.
type RollbackStep interface {
	Step
	Rollback() error
}

// FailurePolicy controls what the runner does after a step fails.
type FailurePolicy int

const (
	// StopOnError stops execution after the first failure (default).
	StopOnError FailurePolicy = iota
	// ContinueOnError runs all steps even if some fail.
	ContinueOnError
)

// RollbackPolicy controls when rollback is triggered.
type RollbackPolicy struct {
	OnApplyFailure bool
}

// DefaultRollbackPolicy returns the standard rollback policy: rollback on any apply failure.
func DefaultRollbackPolicy() RollbackPolicy {
	return RollbackPolicy{OnApplyFailure: true}
}

// ShouldRollback reports whether a rollback should be triggered for stage+error.
func (p RollbackPolicy) ShouldRollback(stage Stage, err error) bool {
	return err != nil && stage == StageApply && p.OnApplyFailure
}

// StagePlan separates the Prepare and Apply steps for a pipeline run.
type StagePlan struct {
	Prepare []Step
	Apply   []Step
}

// StepResult holds the outcome of a single step execution.
type StepResult struct {
	StepID     string
	Status     StepStatus
	Err        error
	StartedAt  time.Time
	FinishedAt time.Time
}

// StageResult holds the outcome of an entire stage.
type StageResult struct {
	Stage   Stage
	Success bool
	Steps   []StepResult
	Err     error
}

// ExecutionResult is the final output of a pipeline run.
type ExecutionResult struct {
	Prepare  StageResult
	Apply    StageResult
	Rollback StageResult
	Err      error
}

// ProgressFunc is called on each step status change during execution.
type ProgressFunc func(ProgressEvent)

// ProgressEvent is emitted for each status transition during execution.
type ProgressEvent struct {
	StepID string
	Stage  Stage
	Status StepStatus
	Err    error
}
