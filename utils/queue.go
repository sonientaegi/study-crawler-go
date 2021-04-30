package utils

import "container/list"

type Queue struct {
	list.List
}

func (q *Queue) Push(v interface{}) int {
	q.PushBack(v)

	return q.Len()
}

func (q *Queue) Pop() interface{} {
	if q.Len() != 0 {
		var e = q.Front()
		q.Remove(e)
		return e.Value
	} else {
		return nil
	}
}

// Init overrides list.init(). Initialize queue and return self pointer.
func (q *Queue) Init() *Queue {
	q.List.Init()
	return q
}
