package system_object

import (
	"fmt"
	"time"
)

type SysObject struct {
	__GLOBAL_OBJECT_ID int
	__created_time     string
	__object_id        int
	__object_id_other  int
}

func (sy *SysObject) __init_() {
	sy.__GLOBAL_OBJECT_ID = 0
	sy.__created_time = time.Now().String()
	sy.__object_id = sy.__GLOBAL_OBJECT_ID
	sy.__GLOBAL_OBJECT_ID = sy.__GLOBAL_OBJECT_ID + 1
	sy.__object_id_other = 0
}

func (sy *SysObject) __str__() string {
	return fmt.Sprintf("ID:%10d %s", sy.__object_id, sy.__created_time)
}

func set_req_time(sy *SysObject) {
	return
}

func get_req_time(sy *SysObject) {
	return
}

func (sy *SysObject) __lt__() bool {
	return sy.__object_id < sy.__object_id_other
}

func (sy *SysObject) get_obj_id() int {
	return sy.__object_id
}