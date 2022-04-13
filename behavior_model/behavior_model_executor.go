package behaviormodel

import (
	"evsim_golang/definition"
	system_object "evsim_golang/system"
	"fmt"
	"math"
)

type BehaviorModelExecutor struct {
	sysobject     *system_object.SysObject
	behaviormodel *Behaviormodel

	_cancel_reshedule_f bool //리스케쥴링펑션의 실행 여부
	engine_name         string
	_instance_t         float64
	_destruct_t         float64
	_cur_state          string
	_next_event_t       float64
	requestedTime       float64
}

func (b *BehaviorModelExecutor) String() string {
	return fmt.Sprintf("[N]:{%s}, [S]:{%s}", b.behaviormodel.coreModel.Get_name(), b._cur_state)
}

func (b *BehaviorModelExecutor) Cancel_rescheduling() {
	b._cancel_reshedule_f = true
}

func (b *BehaviorModelExecutor) Get_engine_name() string {
	return b.engine_name
}

func (b *BehaviorModelExecutor) Set_engine_name(name string) {
	b.engine_name = name
}

func (b *BehaviorModelExecutor) Get_create_time() float64 {
	return b._instance_t
}

func (b *BehaviorModelExecutor) Get_destruct_time() float64 {
	return b._destruct_t
}

func (b *BehaviorModelExecutor) Init_state(state string) {
	b._cur_state = state
}

func (b *BehaviorModelExecutor) Ext_trans(port, msg string) {

}

func (b *BehaviorModelExecutor) Int_trans(port, msg string) {

}
func (b *BehaviorModelExecutor) Output() {

}

func (b *BehaviorModelExecutor) Time_advance() float64 {
	for key, _ := range b.behaviormodel._states {
		if key == b._cur_state {
			return b.behaviormodel._states[b._cur_state]
		}
	}
	return -1
}
func (b *BehaviorModelExecutor) Set_req_time(global_time float64, elapsed_time int) {
	elapsed_time = 0
	if b.Time_advance() == definition.Infinite {
		b._next_event_t = definition.Infinite
		b.requestedTime = definition.Infinite
	} else {
		if b._cancel_reshedule_f {
			b.requestedTime = math.Min(b._next_event_t, global_time+b.Time_advance())
		} else {
			b.requestedTime = global_time + b.Time_advance()
		}
	}
}
func (b *BehaviorModelExecutor) Get_req_time() float64 {
	if b._cancel_reshedule_f {
		b._cancel_reshedule_f = false
	}
	b._next_event_t = b.requestedTime
	return b.requestedTime
}

func NewExecutor(instantiate_time, destruct_time float64, name, engine_name string) *BehaviorModelExecutor {
	b := BehaviorModelExecutor{}
	b.engine_name = engine_name
	b._instance_t = instantiate_time
	b._destruct_t = destruct_time
	b.sysobject = system_object.NewSysObject()
	b.behaviormodel = NewBehaviorModel(name)
	b.requestedTime = math.Inf(1)
	return &b
}
