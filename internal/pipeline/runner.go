package pipeline

import (
	"errors"
	"time"
)

// Runner executes a list of Steps for a given stage.
type Runner struct {
	FailurePolicy FailurePolicy
	OnProgress    ProgressFunc
}

// Run executes all steps in order and returns the StageResult.
// Respects FailurePolicy: StopOnError halts on first failure, ContinueOnError runs all.
func (r Runner) Run(stage Stage, steps []Step) StageResult {
	result := StageResult{
		Stage:   stage,
		Success: true,
		Steps:   make([]StepResult, 0, len(steps)),
	}

	var errs []error

	for _, step := range steps {
		r.emit(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusRunning})

		started := time.Now().UTC()
		err := step.Run()
		finished := time.Now().UTC()

		stepResult := StepResult{
			StepID:     step.ID(),
			StartedAt:  started,
			FinishedAt: finished,
		}

		if err != nil {
			stepResult.Status = StepStatusFailed
			stepResult.Err = err
			result.Steps = append(result.Steps, stepResult)
			r.emit(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusFailed, Err: err})

			errs = append(errs, err)
			result.Success = false

			if r.FailurePolicy == StopOnError {
				result.Err = err
				return result
			}
			continue
		}

		stepResult.Status = StepStatusSucceeded
		result.Steps = append(result.Steps, stepResult)
		r.emit(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusSucceeded})
	}

	if len(errs) > 0 {
		result.Err = errors.Join(errs...)
	}

	return result
}

func (r Runner) emit(event ProgressEvent) {
	if r.OnProgress != nil {
		r.OnProgress(event)
	}
}
