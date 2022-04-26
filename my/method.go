package my

import (
	"evsim_golang/executor"

	"github.com/gammazero/deque"
)

type Pair struct {
	A, B interface{}
}

//슬라이스에서 특정한 값을 찾아 리턴
func Slice_Find(slice []interface{}, val interface{}) (int, bool) {

	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

func Slice_Find_string(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

//맵에서 특정값을 찾음
func Map_Find(m map[interface{}]interface{}, val interface{}) (interface{}, bool) {
	for k, v := range m {
		if v == val {
			return k, true
		}
	}
	return -1, false
}

func Custom_Sorted(list deque.Deque) {
	var A []*executor.BehaviorModelExecutor
	for i := 1; i <= list.Len(); i++ {
		A = append(A, list.PopFront().(*executor.BehaviorModelExecutor))
	}
	for i := 1; i <= list.Len(); i++ {
		for i := list.Len(); i > 0; i-- {
			if A[i].Get_req_time() > A[i-1].Get_req_time() {
				A[i-1], A[i] = A[i], A[i-1]
			}
		}
	}
}
