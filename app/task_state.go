package main

const maxSameResultCnt = 1024

type TaskState struct {
	isFailing     bool
	sameResultCnt int
	everChanged   bool
}

func NewTaskState() *TaskState {
	ts := &TaskState{}
	return ts
}

func (ts *TaskState) Update(isFailed bool) {
	if isFailed != ts.isFailing {
		ts.sameResultCnt = 0
		ts.everChanged = true
	}
	if ts.sameResultCnt <= maxSameResultCnt {
		ts.sameResultCnt++
	}
	ts.isFailing = isFailed
}
