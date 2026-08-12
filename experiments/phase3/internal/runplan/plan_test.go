package runplan

import (
	"testing"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
)

func TestExecutionOrderPrioritizesHighSignalAndRotatesArms(t *testing.T) {
	fixtures := make([]model.Fixture, 0, len(fixturePriority))
	for _, id := range fixturePriority {
		fixtures = append(fixtures, model.Fixture{ID: id})
	}
	arms := make([]model.Arm, 0, len(armOrder))
	for _, id := range armOrder {
		arms = append(arms, model.Arm{ID: id})
	}
	order, err := ExecutionOrder(fixtures, arms)
	if err != nil {
		t.Fatal(err)
	}
	if len(order) != 48 {
		t.Fatalf("unexpected cell count %d", len(order))
	}
	want := []model.CellRef{
		{CaseID: "p3-002", ArmID: "A"},
		{CaseID: "p3-002", ArmID: "B"},
		{CaseID: "p3-008", ArmID: "B"},
		{CaseID: "p3-008", ArmID: "C"},
	}
	for i, expected := range want[:2] {
		if order[i] != expected {
			t.Fatalf("order[%d]=%+v, expected %+v", i, order[i], expected)
		}
	}
	for i, expected := range want[2:] {
		index := len(armOrder) + i
		if order[index] != expected {
			t.Fatalf("order[%d]=%+v, expected %+v", index, order[index], expected)
		}
	}
}
