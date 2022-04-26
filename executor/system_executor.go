package executor

import (
	"container/heap"
	"errors"
	"evsim_golang/definition"
	"evsim_golang/model"
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
	time_step          time.Duration
	EXTERNAL_SRC       string
	EXTERNAL_DST       string
	simulation_mode    int
	min_schedule_item  deque.Deque
	input_event_queue  input_heap
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

type event_queue struct {
	time float64
	msg  interface{}
}

type input_heap []event_queue

func (eq input_heap) Len() int {
	return len(eq)
}

func (eq input_heap) Less(i, j int) bool {
	return false
}

func (eq input_heap) Swap(i, j int) {
	eq[i], eq[j] = eq[j], eq[i]
}

func (eq *input_heap) Push(elem interface{}) {
	*eq = append(*eq, elem.(event_queue))
}

func (eq *input_heap) Pop() interface{} {
	old := *eq
	n := len(old)
	elem := old[n-1]
	*eq = old[0 : n-1]

	return elem
}

//생성자
func NewSysExecutor(_time_step interface{}, _sim_name, _sim_mode string) *SysExecutor {
	se := SysExecutor{}
	se.behaviormodel = model.NewBehaviorModel(_sim_name)
	se.dmc = NewDMC(0, definition.Infinite, "dc", "default")
	se.EXTERNAL_SRC = "SRC"
	se.EXTERNAL_DST = "DST"
	se.global_time = 0
	se.target_time = 0
	se.time_step = _time_step.(time.Duration) * time.Second
	se.simulation_mode = definition.SIMULATION_IDLE
	se.sim_mode = _sim_mode
	se.waiting_obj_map = make(map[float64][]*BehaviorModelExecutor)
	se.active_obj_map = make(map[float64]*BehaviorModelExecutor)
	se.port_map = make(map[Object][]Object)
	se.Register_entity(se.dmc.executor)
	se.min_schedule_item = *deque.New()
	se.output_event_queue = *deque.New()
	se.sim_init_time = time.Now()
	se.input_event_queue = input_heap{}
	heap.Init(&se.input_event_queue)
	return &se
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
		Custom_Sorted(&se.min_schedule_item)

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
			i := se.min_schedule_item.Index(func(i interface{}) bool {
				if i == agent {
					return true
				} else {
					return false
				}
			})
			se.min_schedule_item.Remove(i)
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
		}()
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

func (se *SysExecutor) Init_sim() {
	se.simulation_mode = definition.SIMULATION_RUNNING

	var _del_model map[float64]*BehaviorModelExecutor
	var _del_coupling map[Object]Object

	for _, model_list := range se.waiting_obj_map {
		for _, model := range model_list {
			if model.Behaviormodel.CoreModel.Get_type() == definition.STRUCTURAL {
				// se.Flattening(model, _del_model, _del_coupling) //질문
			}
		}
	}

	for target, _model := range _del_model {
		for _, model := range se.waiting_obj_map[target] {
			if _model == model {
				delete(se.waiting_obj_map, target)
			}
		}
	}

	for target, _model := range _del_coupling {
		for _, model := range se.port_map[target] {
			if _model == model {
				delete(se.port_map, target)
			}
		}
	}

	if !(se.active_obj_map == nil) {
		var min float64 = 0
		for k, _ := range se.waiting_obj_map {
			if k < min {
				min = k
			}
		}
		se.global_time = min
	}

	if !(se.min_schedule_item.Cap() == 0) {
		for _, obj := range se.active_obj_map {
			if obj.Time_advance() < 0 {
				err := func() error {
					return errors.New("You should give posistive real number for the deadline")
				}()
				fmt.Println(err)
			}
			obj.Set_req_time(se.global_time, 0)
			se.min_schedule_item.PushBack(obj)
		}
	}
}
func (se *SysExecutor) Schedule() {
	se.Create_entity()
	se.Handle_external_input_event()

	tuple_obj := se.min_schedule_item.PopFront().(*BehaviorModelExecutor)
	before := time.Now()
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
		se.min_schedule_item.PushFront(tuple_obj)
		Custom_Sorted(&se.min_schedule_item)
		tuple_obj = se.min_schedule_item.PopFront().(*BehaviorModelExecutor)
	}
	se.min_schedule_item.PushFront(tuple_obj)
	after := time.Since(before)
	if se.sim_mode == "REAL_TIME" {

		x := se.time_step - after

		if x < 0 {
			time.Sleep(0)
		} else {
			time.Sleep(x)
		}

	}
	se.global_time += float64(se.time_step)
	se.Destory_entity()

}

func (se *SysExecutor) Simulate(_time float64) { //default = infinity
	se.target_time = se.global_time + _time
	se.Init_sim()

	for se.global_time < se.target_time {
		if se.waiting_obj_map == nil {
			item := se.min_schedule_item.PopFront().(*BehaviorModelExecutor)
			if item.Get_req_time() == definition.Infinite && se.sim_mode == "VIRTURE_TIME" {
				se.simulation_mode = definition.SIMULATION_TERMINATED
				break
			}
		}
		se.Schedule()
	}

}

func (se *SysExecutor) Simulation_stop() {
	se.global_time = 0
	se.target_time = 0
	se.time_step = 1
	se.waiting_obj_map = make(map[float64][]*BehaviorModelExecutor)
	se.active_obj_map = make(map[float64]*BehaviorModelExecutor)
	se.port_map = make(map[Object][]Object)
	se.min_schedule_item = *deque.New()
	se.sim_init_time = time.Now()
	se.dmc = NewDMC(0, definition.Infinite, "dc", "default")
	se.Register_entity(se.dmc.executor)
}

func (se *SysExecutor) Insert_external_event(_port string, _msg interface{}, scheduled_time float64) {
	sm := system.NewSysMessage("SRC", _port)
	sm.Insert(_msg)
	_, bool := Slice_Find_string(se.behaviormodel.CoreModel.Intput_ports, _port)
	if bool == true {
		//lock.acquire
		eq := event_queue{scheduled_time + se.global_time, sm}
		heap.Push(&se.input_event_queue, eq)
		//lock.release()
	} else {
		print("[ERROR][INSERT_EXTERNAL_EVNT] Port Not Found")
	}

}

func (se *SysExecutor) Insert_custom_external_event(_port string, _bodylist []interface{}, scheduled_time float64) {
	sm := system.NewSysMessage("SRC", _port)
	sm.Extend(_bodylist)
	_, bool := Slice_Find_string(se.behaviormodel.CoreModel.Intput_ports, _port)
	if bool == true {
		//lock.acquire
		eq := event_queue{scheduled_time + se.global_time, sm}
		heap.Push(&se.input_event_queue, eq)
		//lock.release()
	} else {
		fmt.Printf("[ERROR][INSERT_EXTERNAL_EVNT] Port Not Found")
	}
}

func (se *SysExecutor) Get_generated_event() deque.Deque {
	return se.output_event_queue
}

func (se *SysExecutor) Handle_external_input_event() {
	var event_list []event_queue
	for _, ev := range se.input_event_queue {
		if ev.time <= se.global_time {
			event_list = append(event_list, ev)
		}
	}
	for _, event := range event_list {
		se.output_handling(nil, event)
		heap.Pop(&se.input_event_queue)
	}
	Custom_Sorted(&se.min_schedule_item)
}

func (se *SysExecutor) Handle_external_output_event() deque.Deque {
	var event_lists deque.Deque
	err := deepcopy.Copy(event_lists, se.output_event_queue)
	if err != nil {
		err := func() error {
			return errors.New("can't Handle_external_output_event")
		}()
		fmt.Println(err)
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
