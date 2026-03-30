package pipeline

// OrchestratorOption configures the Orchestrator.
type OrchestratorOption func(*Orchestrator)

// WithFailurePolicy sets the failure policy for the apply stage runner.
func WithFailurePolicy(policy FailurePolicy) OrchestratorOption {
	return func(o *Orchestrator) {
		o.runner.FailurePolicy = policy
	}
}

// WithProgressFunc sets a callback that receives progress events during execution.
func WithProgressFunc(fn ProgressFunc) OrchestratorOption {
	return func(o *Orchestrator) {
		o.runner.OnProgress = fn
	}
}

// Orchestrator runs a 2-stage pipeline (Prepare → Apply) with automatic rollback on failure.
type Orchestrator struct {
	runner   Runner
	policy   RollbackPolicy
	stepByID map[string]Step
}

// NewOrchestrator creates an Orchestrator with the given rollback policy and options.
func NewOrchestrator(policy RollbackPolicy, opts ...OrchestratorOption) *Orchestrator {
	o := &Orchestrator{
		runner:   Runner{FailurePolicy: StopOnError},
		policy:   policy,
		stepByID: map[string]Step{},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Execute runs the StagePlan through Prepare then Apply.
// If Apply fails and the rollback policy allows it, completed Apply steps are rolled back.
func (o *Orchestrator) Execute(plan StagePlan) ExecutionResult {
	o.indexSteps(plan.Prepare)
	o.indexSteps(plan.Apply)

	prepareResult := o.runner.Run(StagePrepare, plan.Prepare)
	if !prepareResult.Success {
		return ExecutionResult{Prepare: prepareResult, Err: prepareResult.Err}
	}

	applyResult := o.runner.Run(StageApply, plan.Apply)
	result := ExecutionResult{Prepare: prepareResult, Apply: applyResult}
	if applyResult.Success {
		return result
	}

	result.Err = applyResult.Err
	if o.policy.ShouldRollback(StageApply, applyResult.Err) {
		result.Rollback = ExecuteRollback(applyResult.Steps, o.stepByID)
		if !result.Rollback.Success {
			result.Err = result.Rollback.Err
		}
	}

	return result
}

func (o *Orchestrator) indexSteps(steps []Step) {
	for _, step := range steps {
		o.stepByID[step.ID()] = step
	}
}
