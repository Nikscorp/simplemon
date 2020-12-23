package main

import (
	"context"
	"log"
	"sync"
	"time"
)

const NotifierMaxTimeoutSec = 60

type Notifier interface {
	NotifyOrIgnore(context.Context, TaskResult, map[string]TaskState)
}

type NotifiersManager struct {
	Queue      chan TaskResult
	TasksState map[string]TaskState
	Notifiers  []Notifier
}

func NewNotifiersManager(queue chan TaskResult) NotifiersManager {
	nm := NotifiersManager{Queue: queue}
	nm.TasksState = make(map[string]TaskState)
	nm.Notifiers = make([]Notifier, 0)
	return nm
}

func (nm *NotifiersManager) RegisterNotifier(n Notifier) {
	nm.Notifiers = append(nm.Notifiers, n)
}

func (nm *NotifiersManager) RunNotifiersManager(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			break
		case tr := <-nm.Queue:
			nm.FixState(tr)
			nm.NotifyOrIgnoreAll(ctx, tr)
		}
	}
}

func (nm *NotifiersManager) FixState(tr TaskResult) {
	_, isPresent := nm.TasksState[tr.Task.ID]
	if !isPresent {
		nm.TasksState[tr.Task.ID] = *NewTaskState()
	}
	state := nm.TasksState[tr.Task.ID]
	state.Update(tr.WasError)
	nm.TasksState[tr.Task.ID] = state
}

func (nm *NotifiersManager) NotifyOrIgnoreAll(ctx context.Context, tr TaskResult) {
	newCtx, cancel := context.WithTimeout(ctx, NotifierMaxTimeoutSec*time.Second)
	var wg sync.WaitGroup
	for _, n := range nm.Notifiers {
		wg.Add(1)
		go func(n Notifier) {
			n.NotifyOrIgnore(newCtx, tr, nm.TasksState)
			wg.Done()
		}(n)
	}
	wg.Wait()
	cancel()
	log.Printf("[INFO] Notify done for task %s", tr.Task.ID)
}
