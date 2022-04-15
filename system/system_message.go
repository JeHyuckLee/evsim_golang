package system_object

type SysMessage struct {
	_src      string
	_dst      string
	_msg_time float64
	_msg_list []string
}

func NewSysMessage(src_name string, dst_name string) *SysMessage {
	sy := SysMessage{}
	sy._src = src_name
	sy._dst = dst_name
	sy._msg_time = -1
	return &sy
}

func (b *SysMessage) Insert(msg string) {
	b._msg_list = append(b._msg_list, msg)
}

func (b *SysMessage) Extend(_list []string) {
	b._msg_list = append(b._msg_list, _list...)
}

func (b *SysMessage) Retrieve() []string {
	return b._msg_list
}

func (b *SysMessage) Get_src() string {
	return b._src
}

func (b *SysMessage) Get_dst() string {
	return b._dst
}

func (b *SysMessage) Set_msg_time(t float64) {
	b._msg_time = t
}

func (b *SysMessage) Get_msg_time() float64 {
	return b._msg_time
}
