package main

import (
	"evsim_golang/definition"
	"evsim_golang/executor"
	"evsim_golang/system"
	"fmt"
	"time"
)

type Generator struct {
	executor *executor.BehaviorModelExecutor
	msg_list []interface{}
}

func (g *Generator) Ext_trans(port string, msg *system.SysMessage) {
	fmt.Println("ext_trans")
	if port == "start" {
		fmt.Println("[gen][out]:", time.Now())
		g.executor.Cur_state = "MOVE"
	}
}

func (g *Generator) Int_trans() {
	if g.executor.Cur_state == "SEND" && g.msg_list == nil {
		g.executor.Cur_state = "IDLE"
	} else {
		g.executor.Cur_state = "SEND"
	}
}

func (g *Generator) Output() *system.SysMessage {
	fmt.Println("output")
	msg := system.NewSysMessage(g.executor.Behaviormodel.CoreModel.Get_name(), "process")
	fmt.Println("[gen][out]:", time.Now())
	msg.Insert(g.msg_list[0])
	g.msg_list = remove(g.msg_list, 0)

	return msg
}

func NewGenerator() *Generator {
	gen := Generator{}
	gen.executor = executor.NewExecutor(0, definition.Infinite, "Gen", "sname")
	gen.executor.Init_state("IDLE")
	gen.executor.Behaviormodel.Insert_state("IDLE", definition.Infinite)
	gen.executor.Behaviormodel.Insert_state("SEND", 1)
	gen.executor.Behaviormodel.Insert_state("MOVE", 1)
	gen.executor.Behaviormodel.CoreModel.Insert_input_port("start")
	gen.executor.Behaviormodel.CoreModel.Insert_output_port("process")
	gen.msg_list = append(gen.msg_list, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10)
	return &gen
}

type Processor struct {
	executor *executor.BehaviorModelExecutor
	msg_list []interface{}
}

func (p *Processor) Ext_trans(port string, msg *system.SysMessage) {
	if port == "process" {
		fmt.Println("[proc][in]", time.Now())
		p.executor.Cancel_rescheduling()
		data := msg.Retrieve()
		p.msg_list = append(p.msg_list, data...)
		p.executor.Cur_state = "PROCESS"

	}
}

func (p *Processor) Int_trans() {
	if p.executor.Cur_state == "PROCESS" {
		p.executor.Cur_state = "IDLE"
	} else {
		p.executor.Cur_state = "IDLE"
	}
}

func (p Processor) Output() *system.SysMessage {
	fmt.Println("[proc][out]", time.Now())
	fmt.Println(p.msg_list...)
	return nil
}

func NewProcessor() *Processor {
	pro := Processor{}
	pro.executor = executor.NewExecutor(0, definition.Infinite, "Proc", "sname")
	pro.executor.Init_state("IDLE")
	pro.executor.Behaviormodel.Insert_state("IDLE", definition.Infinite)
	pro.executor.Behaviormodel.Insert_state("PROCESS", 2)
	pro.executor.Behaviormodel.CoreModel.Insert_input_port("PROCESS")

	return &pro
}

func main() {
	se := NewSysSimulator()
	se.Register_engine("sname", "REAL_TIME", 1)
	sim := se.Get_engine("sname")
	sim.Behaviormodel.CoreModel.Insert_input_port("start")

	gen := NewGenerator()
	pro := NewProcessor()

	sim.Register_entity(gen.executor)

	sim.Register_entity(gen.executor)

	sim.Coupling_relation(nil, "strart", gen.executor, "start")
	sim.Coupling_relation(gen.executor, "process", pro.executor, "process")
	sim.Insert_external_event("start", nil, 0)
	sim.Simulate(definition.Infinite)

}

func remove(slice []interface{}, s int) []interface{} {
	return append(slice[:s], slice[s+1:]...)
}
