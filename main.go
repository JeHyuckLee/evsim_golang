package main

import (
	"evsim_golang/definition"
	"evsim_golang/executor"
	"evsim_golang/system"
	"fmt"
)

type Generator struct {
	executor *executor.BehaviorModelExecutor
	msg_list []interface{}
}

func NewGenerator(instance_time, destruct_time float64, name, engine_name string) *Generator {
	gen := Generator{}
	gen.executor = executor.NewExecutor(instance_time, destruct_time, name, engine_name)
	gen.executor.AbstractModel = &gen
	gen.executor.Init_state("IDLE")
	gen.executor.Behaviormodel.Insert_state("IDLE", definition.Infinite)
	gen.executor.Behaviormodel.Insert_state("MOVE", 1)
	gen.executor.Behaviormodel.CoreModel.Insert_input_port("start")
	gen.executor.Behaviormodel.CoreModel.Insert_output_port("process")
	for i := 0; i < 10; i++ {
		gen.msg_list = append(gen.msg_list, i)
	}
	return &gen
}

func (g *Generator) Ext_trans(port string, msg *system.SysMessage) {

	//fmt.Println("ext_trans")
	if port == "start" {
		g.executor.Cur_state = "MOVE"
	}
}

func (g *Generator) Int_trans() {
	//fmt.Println("int_trans")
	if g.executor.Cur_state == "MOVE" && g.msg_list == nil {
		g.executor.Cur_state = "IDLE"
	} else {
		g.executor.Cur_state = "MOVE"
	}
}

func (g *Generator) Output() *system.SysMessage {

	fmt.Println(g.msg_list...)
	g.msg_list = remove(g.msg_list, 0)
	return nil
}

func main() {
	se := executor.NewSysSimulator()            // 엔진 생성
	se.Register_engine("sname", "REAL_TIME", 1) // 엔진 등록
	sim := se.Get_engine("sname")               // 엔진 가져오기

	sim.Behaviormodel.CoreModel.Insert_input_port("start") // 입력 포트 생성

	gen := NewGenerator(0, definition.Infinite, "Gen", "sname") // 정의한 오브젝트 가져오기
	// 오브젝트의 생성시간, 파괴시간, 이름, 등록할 엔진이름

	sim.Register_entity(gen.executor) // 오브젝트 등록

	sim.Coupling_relation(nil, "start", gen.executor, "start") // 엔진 포트와 오브젝트 포트 연결

	sim.Insert_external_event("start", nil, 0) // 발생시킬 이벤트 정의

	sim.Simulate(definition.Infinite) // 엔진 실행
}

func remove(slice []interface{}, s int) []interface{} {
	return append(slice[:s], slice[s+1:]...)
}
