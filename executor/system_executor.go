package executor

import (
	"errors"
	"evsim_golang/definition"
	"evsim_golang/model"
	"evsim_golang/my"
	"evsim_golang/system"
	"fmt"
	"math"
	"time"

	"github.com/gammazero/deque"
	"gopkg.in/getlantern/deepcopy.v1"
)

type SysExecutor struct {
	sysObject     *system.SysObject
	behaviormodel *model.Behaviormodel
	dmc           *DefaultMessageCatcher

	global_time        float64
	target_time        float64
	time_step          int
	EXTERNAL_SRC       string
	EXTERNAL_DST       string
	simulation_mode    int
	min_schedule_item  deque.Deque
	input_event_queue  []string
	output_event_queue deque.Deque
	sim_mode           string
	waiting_obj_map    map[float64][]*BehaviorModelExecutor
	active_obj_map     map[float64]*BehaviorModelExecutor
	learn_module       interface{}
	port_map           map[Object][]Object
	sim_init_time      time.Time
}

type Object struct {
	object *BehaviorModelExecutor
	port   string
}

func NewSysExecutor(_time_step int, _sim_name, _sim_mode string) *SysExecutor {
	se := SysExecutor{}
	se.behaviormodel = model.NewBehaviorModel(_sim_name)
	se.dmc = NewDMC(0, definition.Infinite, "dc", "default")
	se.EXTERNAL_SRC = "SRC"
	se.EXTERNAL_DST = "DST"
	se.global_time = 0
	se.target_time = 0
	se.time_step = _time_step
	se.simulation_mode = definition.SIMULATION_IDLE
	se.sim_mode = _sim_mode
	se.waiting_obj_map = make(map[float64][]*BehaviorModelExecutor)
	se.active_obj_map = make(map[float64]*BehaviorModelExecutor)
	se.port_map = make(map[Object][]Object)
	se.Register_entity(se.dmc.executor)
	se.min_schedule_item = *deque.New()
	se.output_event_queue = *deque.New()
	se.sim_init_time = time.Now()
}

func (se SysExecutor) Get_global_time() float64 {
	return se.global_time
}

func (se *SysExecutor) Register_entity(sim_obj *BehaviorModelExecutor) {
	se.waiting_obj_map[sim_obj.Get_create_time()] = append(se.waiting_obj_map[sim_obj.Get_create_time()], sim_obj)
	// waiting_obj_map 에 create_time 별로 슬라이스를 만들어서 sim_obj 를 append 한다.
}

func (se *SysExecutor) Create_entity() {
	if len(se.waiting_obj_map) != 0 {
		key, value := func() (float64, []*BehaviorModelExecutor) {
			var key float64 = 0
			for k, _ := range se.waiting_obj_map {
				if k < key {
					key = k
				}
			}
			value := se.waiting_obj_map[key]
			return key, value
		}() //key = create_time, value = obj의 슬라이스
		for _, v := range value {
			se.active_obj_map[float64(v.sysobject.Get_obj_id())] = v
			v.Set_req_time(se.global_time, 0) //elpased ti
			se.min_schedule_item.PushFront(v)
			//슬라이스를 순회하여 obj 를 active_obj_map 에 넣는다.
		}
		delete(se.waiting_obj_map, key)
		// se.min_schedule_itme 정렬
		my.Custom_Sorted(se.min_schedule_item)
	}
}

func (se *SysExecutor) Destory_entity() {
	if len(se.active_obj_map) != 0 { //active obj map 에 obj 가 있으면
		var delete_lst []*BehaviorModelExecutor
		// var port_del_lst []string
		for _, agent := range se.active_obj_map {
			if agent.Get_create_time() <= se.global_time { //active_obj_map을 순회하고,
				delete_lst = append(delete_lst, agent) // 이미생성된 obj 들을 delete_lst에 담고
			}
		}
		for _, agent := range delete_lst {
			delete(se.active_obj_map, float64(agent.sysobject.Get_obj_id())) // delete_lst를 순회하여 active_obj_map 에 있는 obj 를 지운다.
			var port_del_lst []Object
			for k, v := range se.port_map {
				if v[0].object == agent { //지운 obj 와 연결되어있는 port 를 port_map에서 지운다.
					port_del_lst = append(port_del_lst, k)
				}
			}
			for _, v := range port_del_lst {
				delete(se.port_map, v)
			}
			//se.min_schedule_item.Remove(agent)
			//mim_schedule_item에서도 지운다.
		}
	}
}

func (se *SysExecutor) Coupling_relation(src_obj, dst_obj *BehaviorModelExecutor, out_port, in_port string) {
	dst := Object{dst_obj, in_port}
	b := func() bool {
		for k, _ := range se.port_map {
			if k.object == src_obj && k.port == out_port {
				se.port_map[k] = append(se.port_map[k], dst)
				return true //port_map 에 이미있으면 추가
			}
		}
		return false
	}()
	if b == false { // 없으면 새로만든다.
		src := Object{src_obj, out_port}
		se.port_map[src] = append(se.port_map[src], dst)
	}
}

// func (se *SysExecutor) _Coupling_relation(src, dst interface{}) {
// 	_, bool := my.Map_Find(se.port_map, src)
// 	if bool == true {
// 		se.port_map[src] = dst
// 	} else {
// 		se.port_map[src] = dst
// 	}
// // }
// type pair struct {
// 	pair_object *obj
// 	dst         string
// }
type event_queue struct {
	time float64
	msg  interface{}
}

func (se *SysExecutor) Single_output_handling(obj *BehaviorModelExecutor, msg *system.SysMessage) {
	pair := Object{obj, msg.Get_dst()}

	b := func() bool {
		for k, _ := range se.port_map {
			if k.object == obj {
				return true
			}
		}
		return false
	}()
	if b == false {
		dmc := Object{se.active_obj_map[float64(se.dmc.executor.sysobject.Get_obj_id())], "uncaught"}
		se.port_map[pair] = append(se.port_map[pair], dmc)
	}

	dst := se.port_map[pair]
	if dst == nil { //도착지가없다
		err := func() error {
			return errors.New("Destination Not Found")
		}
		fmt.Println(err)
	}
	for _, v := range dst {
		if v.object == nil {
			e := event_queue{se.global_time, msg.Retrieve()}
			se.output_event_queue.PushFront(e)
		} else {
			v.object.Ext_trans(v.port, msg.Retrieve())
			v.object.Set_req_time(se.global_time, 0)
		}
	}
}

func (se *SysExecutor) output_handling(obj, msg interface{}) {
	if !(msg == nil) {
		// if type(msg) == list:
		//         for ith_msg in msg:
		//             self.single_output_handling(obj, copy.deepcopy(ith_msg))
		//     else:
		//         self.single_output_handling(obj, msg)
	}
}

func (se *SysExecutor) Flattening(_model []string, _del_model []string, _del_coupling []string) bool {
	for k, v := range _model.retrieve_external_output_coupling { //  핸들 외부 출력 커플링
		if _, v := se.port_map[k]; v {
			for coupling := range se.port_map[v] {
				se._coupling_relation(k, coupling)
				_del_coupling = append(_del_coupling, v, coupling)
			}
		}
	}
	for k, v := range _model.retrieve_external_input_coupling { // 핸들 외부 입력 커플링
		var port_key_lst []
		for sk, sv := range se.port_map {
			if k, _ := sv; _ {
				port_key_lst = append(port_key_lst, sk)
				del_coupling = append(del_coupling, sk, k)
			}
			for key := range port_key_lst {
				se.port_map[key].extend(v) // extend - iterable 자료형 -> 변환가능한 객체만 가능
			}
		}
	}
	for k, v := range _model.retrieve_internal_coupling { // 핸들 내부 커플링
		for dst := range v {
			se._coupling_relation(k, dst)
		}
	}
	for m := range _model.retrieve_models { // 모델을 계층적으로 관리
		if m.get_type() == ModelType.STRUCTURAL {
			se.flattening(m, _del_model, _del_coupling)
		} else {
			se.register_entity(m)
		}
	}
	for k, model_lst := range self.waiting_obj_map {
		// if _model in model_lst{
		// 	_del_model = append(_del_model, k, _model)
		// }
	}
}

func (se *SysExecutor) Init_sim() {
	se.simulation_mode = definition.SIMULATION_RUNNING

	var _del_model []*BehaviorModelExecutor
	var _del_coupling []*BehaviorModelExecutor

	for _, model_list := range se.waiting_obj_map {
		for _, model := range model_list {
			if model.Get_type() == definition.BEHAVIORAL {
				se.Flattening(model, _del_model, _del_coupling)
			}
		}
	}

	// for target, _model := range _del_model {
	// 	if _model := se.waiting_obj_map[target] {
	// 		se.waiting_obj_map[target].remove(_model)
	// 	}
	// }

	// for target, _model := range _del_coupling {
	// 	if _model := se.port_map[target] {
	// 		se.port_map[target].remove(_model)
	// 	}
	// }

	// if !(se.active_obj_map == nil) { se.global_time = my.Min(se.waiting_obj_map) }

	// if !(se.min_schedule_item) {
	// 	for obj := se.active_obj_map.Items(){
	// 		if obj[1].Time_advance() < 0 {
	// 			print("You should give posistive real number for the deadline")
	// 		}
	// 		obj[1].Set_req_time(se.global_time)
	// 		se.min_schedule_item = append(se.min_schedule_item, obj[1])
	// 	}
}

func (se *SysExecutor) Schedule() {
	se.Create_entity()
	se.Handle_external_input_event()

	tuple_obj := se.min_schedule_item.PopFront().(*BehaviorModelExecutor)

	for {
		t := math.Abs(tuple_obj.Get_req_time() - se.global_time) //req_time 과 global time 의 오차가 1e-9 보다 작으면 true
		if t > 1e-9 {
			break
		}
		msg := tuple_obj.Output()
		if msg != nil {
			// self.output_handling(tuple_obj, (self.global_time, msg))
		}
		// tuple_obj.Int_trans()
		req_t := tuple_obj.Get_req_time()
		tuple_obj.Set_req_time(req_t, 0)

	}

	se.global_time += float64(se.time_step)
	se.Destory_entity()

}
func (se *SysExecutor) Simulate(_time float64) { //default = infinity
	se.target_time = se.global_time + _time
	se.Init_sim()

	for se.global_time < se.target_time {
		// if not self.waiting_obj_map:
		// if self.min_schedule_item[0].get_req_time(
		// ) == Infinite and self.sim_mode == 'VIRTUAL_TIME':
		// 	self.simulation_mode = SimulationMode.SIMULATION_TERMINATED

		se.Schedule()
	}

}

func (se *SysExecutor) Simulation_stop() {
	se.global_time = 0
	se.target_time = 0
	se.time_step = 1
	for k := range se.waiting_obj_map {
		delete(se.waiting_obj_map, k)
	}
	for k := range se.active_obj_map {
		delete(se.active_obj_map, k)
	}
	for k := range se.port_map {
		delete(se.port_map, k)
	}
	se.min_schedule_item = *deque.New()
	se.sim_init_time = time.Now()
	se.dmc = NewDMC(0, definition.Infinite, "dc", "default")
	se.Register_entity(se.dmc.executor)

}

func (se *SysExecutor) Insert_external_event(_port, _msg string, scheduled_time int) {
	sm := system.NewSysMessage("SRC", _port)
	sm.Insert(_msg)
	_, bool := my.Slice_Find_string(se.behaviormodel.CoreModel.Intput_ports, _port)
	if bool == true {
		// self.lock.acquire()
		// heapq.heappush(self.input_event_queue,
		//                (scheduled_time + self.global_time, sm))
		// self.lock.release()
	} else {
		print("[ERROR][INSERT_EXTERNAL_EVNT] Port Not Found")
	}

}

func (se *SysExecutor) Insert_custom_external_event(_port string, _bodylist []string, scheduled_time interface{}) {
	sm := system.NewSysMessage("SRC", _port)
	sm.Extend(_bodylist)
	_, bool := my.Slice_Find_string(se.behaviormodel.CoreModel.Intput_ports, _port)
	if bool == true {
		// self.lock.acquire()
		// heapq.heappush(self.input_event_queue,
		//                (scheduled_time + self.global_time, sm))
		// self.lock.release()
	} else {
		fmt.Printf("[ERROR][INSERT_EXTERNAL_EVNT] Port Not Found")
	}
}

func (se *SysExecutor) Get_generated_event() deque.Deque {
	return se.output_event_queue
}

func (se *SysExecutor) Handle_external_input_event() {
	// event_list = [
	//         ev for ev in self.input_event_queue if ev[0] <= self.global_time
	//     ]
	//     #print(event_list)
	//     for event in event_list:
	//         self.output_handling(None, event)
	//         self.lock.acquire()
	//         heapq.heappop(self.input_event_queue)
	//         self.lock.release()

	//     self.min_schedule_item = deque(
	//         sorted(self.min_schedule_item, key=lambda bm: bm.get_req_time()))
	//     pass
}

func (se *SysExecutor) Handle_external_output_event() deque.Deque {
	var event_lists deque.Deque
	err := deepcopy.Copy(event_lists, se.output_event_queue)
	if err != nil {

	}
	se.output_event_queue.Clear()
	return event_lists
}

func (se *SysExecutor) Is_terminated() interface{} {
	return se.simulation_mode == definition.SIMULATION_TERMINATED
}

func (se SysExecutor) Set_learning_module(learn_module interface{}) {
	se.learn_module = learn_module
}

func (se *SysExecutor) Get_learning_module() interface{} {
	return se.learn_module
}