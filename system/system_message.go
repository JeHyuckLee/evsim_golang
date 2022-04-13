package system

type SysMessage struct {
	_src      string
	_dst      string
	_msg_time float64
	_msg_list []string
}

func (b *SysMessage) String() string {
	return
}

func (b *SysMessage) insert(msg string) {
	b._msg_list = append(b._msg_list, msg)
}

func (b *SysMessage) extend(_list []string) {
	b._msg_list = extend(b._msg_list, _list)
}

func (b *SysMessage) retrieve() []string {
	return b._msg_list
}

func (b *SysMessage) get_src() string {
	return b._src
}

func (b *SysMessage) get_dst() string {
	return b._dst
}

func (b *SysMessage) set_msg_time(t float64) {
	b._msg_time = t
}

func (b *SysMessage) get_msg_time() float64 {
	return b._msg_time
}
