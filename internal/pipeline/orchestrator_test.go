package pipeline_test

import (
	"errors"
	"testing"

	"github.com/ismartz/aispace-setup/internal/pipeline"
)

// stubStep is a minimal Step implementation for testing.
type stubStep struct {
	id      string
	runErr  error
	ran     bool
}

func (s *stubStep) ID() string  { return s.id }
func (s *stubStep) Run() error  { s.ran = true; return s.runErr }

// rollbackStubStep is a Step that also supports rollback.
type rollbackStubStep struct {
	stubStep
	rollbackErr  error
	rolledBack   bool
}

func (s *rollbackStubStep) Rollback() error { s.rolledBack = true; return s.rollbackErr }

func TestOrchestrator_success(t *testing.T) {
	a := &stubStep{id: "prepare-snapshot"}
	b := &rollbackStubStep{stubStep: stubStep{id: "apply-persona"}}
	c := &rollbackStubStep{stubStep: stubStep{id: "apply-sdd"}}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(pipeline.StagePlan{
		Prepare: []pipeline.Step{a},
		Apply:   []pipeline.Step{b, c},
	})

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Apply.Success {
		t.Error("apply stage should succeed")
	}
	if b.rolledBack || c.rolledBack {
		t.Error("no rollback should happen on success")
	}
}

func TestOrchestrator_rollbackOnApplyFailure(t *testing.T) {
	failErr := errors.New("inject failed")

	a := &rollbackStubStep{stubStep: stubStep{id: "apply-persona"}}
	b := &rollbackStubStep{stubStep: stubStep{id: "apply-sdd", runErr: failErr}}
	c := &rollbackStubStep{stubStep: stubStep{id: "apply-mcp"}}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(pipeline.StagePlan{
		Apply: []pipeline.Step{a, b, c},
	})

	if result.Err == nil {
		t.Fatal("expected error")
	}
	if !a.rolledBack {
		t.Error("step 'a' (completed before failure) should be rolled back")
	}
	if c.ran {
		t.Error("step 'c' should not have run after failure")
	}
}

func TestOrchestrator_prepareFailure_noApply(t *testing.T) {
	prepErr := errors.New("backup failed")
	prep := &stubStep{id: "backup", runErr: prepErr}
	apply := &rollbackStubStep{stubStep: stubStep{id: "apply-persona"}}

	orch := pipeline.NewOrchestrator(pipeline.DefaultRollbackPolicy())
	result := orch.Execute(pipeline.StagePlan{
		Prepare: []pipeline.Step{prep},
		Apply:   []pipeline.Step{apply},
	})

	if result.Err == nil {
		t.Fatal("expected error from prepare failure")
	}
	if apply.ran {
		t.Error("apply should not run if prepare fails")
	}
}
