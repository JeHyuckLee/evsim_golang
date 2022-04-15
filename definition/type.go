package definition

import (
	"math"
	"strings"
)

type AttributeType struct {
	aType map[string]int
}

var Infinite float64 = math.Inf(1) //양의 무한대
<<<<<<< HEAD
=======

const (
	SIMULATION_IDLE       = 0
	SIMULATION_RUNNING    = 1
	SIMULATION_TERMINATED = 2
	SIMULATION_PAUSE      = 3
	SIMULATION_UNKNOWN    = -1
>>>>>>> 138dbac8a7345d6ca6951e68c1873c1665d0ebb0

const ( //SimulationMode
	SIMULATION_IDLE       = iota
	SIMULATION_RUNNING    = iota
	SIMULATION_TERMINATED = iota
	SIMULATION_PAUSE      = iota
	SIMULATION_UNKNOWN    = -1
)

const ( //ModelType
	BEHAVIORAL = iota
	STRUCTURAL = iota
)

func (a AttributeType) Resolve_type_form_str(name string) int {
	if "ASPECT" == strings.ToUpper(name) {
		return a.aType["ASPECT"]
	} else if "RUNTIME" == strings.ToUpper(name) {
		return a.aType["RUNTIME"]
	} else {
		return a.aType["UNKNOWN_TYPE"]
	}
}

func (a AttributeType) Resolve_type_from_enum(enum int) string {
	if enum == a.aType["ASPECT"] {
		return "ASPECT"
	} else if enum == a.aType["RUNTIME"] {
		return "RUNTIME"
	} else {
		return "UNKNOWN"
	}
}

func NewAttributeType() *AttributeType {
	a := AttributeType{}
	a.aType["ASPECT"] = 1
	a.aType["RUNTIME"] = 2
	a.aType["UNKNOWN_TYPE"] = -1
	return &a
}
