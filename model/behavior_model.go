package model

import (
	"evsim_golang/definition"
	"math"
)

type Behaviormodel struct {
	States                        map[string]float64
	external_transition_map_tuple []string
	external_transition_map_state []string
	internal_transition_map_tuple []string
	internal_transition_map_state []string
	CoreModel                     *definition.CoreModel
}

func (b *Behaviormodel) Insert_state(name string, deadline float64) { //deadline 디폴트값 = 0
	if deadline == 0 {
		deadline = math.Inf(1)
	}
	b.States[name] = deadline
}

func (b *Behaviormodel) Update_state(name string, deadline float64) { //deadline 디폴트값 = 0
	if deadline == 0 {
		deadline = math.Inf(1)
	}
	b.States[name] = deadline
}

func NewBehaviorModel(name string) *Behaviormodel {
	b := Behaviormodel{}
	b.CoreModel = definition.NewCoreModel(name, definition.BEHAVIORAL)
	return &b
}
