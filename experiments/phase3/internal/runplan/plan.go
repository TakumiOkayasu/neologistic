package runplan

import (
	"fmt"

	"github.com/TakumiOkayasu/neologistic/experiments/phase3/internal/model"
)

var fixturePriority = []string{
	"p3-002",
	"p3-008",
	"p3-003",
	"p3-005",
	"p3-006",
	"p3-007",
	"p3-004",
	"p3-001",
}

var armOrder = []string{"A", "B", "C", "D", "E", "F"}

func ExecutionOrder(fixtures []model.Fixture, arms []model.Arm) ([]model.CellRef, error) {
	fixtureSet := make(map[string]struct{}, len(fixtures))
	for _, fixture := range fixtures {
		fixtureSet[fixture.ID] = struct{}{}
	}
	armSet := make(map[string]struct{}, len(arms))
	for _, arm := range arms {
		armSet[arm.ID] = struct{}{}
	}
	if len(fixtureSet) != len(fixturePriority) || len(armSet) != len(armOrder) {
		return nil, fmt.Errorf("run plan requires canonical fixture and arm sets")
	}
	for _, fixtureID := range fixturePriority {
		if _, ok := fixtureSet[fixtureID]; !ok {
			return nil, fmt.Errorf("run plan fixture %s is missing", fixtureID)
		}
	}
	for _, armID := range armOrder {
		if _, ok := armSet[armID]; !ok {
			return nil, fmt.Errorf("run plan arm %s is missing", armID)
		}
	}

	order := make([]model.CellRef, 0, len(fixturePriority)*len(armOrder))
	for fixtureIndex, fixtureID := range fixturePriority {
		for offset := range armOrder {
			armID := armOrder[(fixtureIndex+offset)%len(armOrder)]
			order = append(order, model.CellRef{CaseID: fixtureID, ArmID: armID})
		}
	}
	return order, nil
}
