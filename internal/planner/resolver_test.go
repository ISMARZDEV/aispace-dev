package planner_test

import (
	"testing"

	"github.com/ismartz/aispace-setup/internal/model"
	"github.com/ismartz/aispace-setup/internal/planner"
)

func TestResolve_autoAddsDependencies(t *testing.T) {
	resolver := planner.NewResolver(planner.DefaultGraph())

	// User selects SDD — should auto-add Engram.
	selection := model.Selection{
		Components: []model.ComponentID{model.ComponentSDD},
	}
	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(plan.OrderedComponents, model.ComponentEngram) {
		t.Error("expected Engram in ordered components")
	}
	if !contains(plan.AddedDependencies, model.ComponentEngram) {
		t.Error("expected Engram in added dependencies")
	}
}

func TestResolve_orderEngramBeforeSDD(t *testing.T) {
	resolver := planner.NewResolver(planner.DefaultGraph())

	selection := model.Selection{
		Components: []model.ComponentID{model.ComponentSDD},
	}
	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	engramIdx := indexOf(plan.OrderedComponents, model.ComponentEngram)
	sddIdx := indexOf(plan.OrderedComponents, model.ComponentSDD)

	if engramIdx == -1 || sddIdx == -1 {
		t.Fatal("missing components in plan")
	}
	if engramIdx >= sddIdx {
		t.Errorf("Engram (%d) should come before SDD (%d)", engramIdx, sddIdx)
	}
}

func TestResolve_personaBeforeEngram_softOrdering(t *testing.T) {
	resolver := planner.NewResolver(planner.DefaultGraph())

	selection := model.Selection{
		Components: []model.ComponentID{
			model.ComponentEngram,
			model.ComponentPersona,
		},
	}
	plan, err := resolver.Resolve(selection)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	personaIdx := indexOf(plan.OrderedComponents, model.ComponentPersona)
	engramIdx := indexOf(plan.OrderedComponents, model.ComponentEngram)

	if personaIdx == -1 || engramIdx == -1 {
		t.Fatal("missing components")
	}
	if personaIdx >= engramIdx {
		t.Errorf("Persona (%d) should come before Engram (%d)", personaIdx, engramIdx)
	}
}

func TestResolve_unknownComponent(t *testing.T) {
	resolver := planner.NewResolver(planner.DefaultGraph())

	selection := model.Selection{
		Components: []model.ComponentID{"nonexistent"},
	}
	_, err := resolver.Resolve(selection)
	if err == nil {
		t.Error("expected error for unknown component")
	}
}

func contains(slice []model.ComponentID, item model.ComponentID) bool {
	for _, c := range slice {
		if c == item {
			return true
		}
	}
	return false
}

func indexOf(slice []model.ComponentID, item model.ComponentID) int {
	for i, c := range slice {
		if c == item {
			return i
		}
	}
	return -1
}
