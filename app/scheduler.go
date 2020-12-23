package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/google/shlex"
)

const DefaultChannelSize = 100

type Scheduler struct {
	Config           Config
	TasksQueue       chan Task
	TasksResultQueue chan TaskResult
	NotifiersManager NotifiersManager
}

type TaskResult struct {
	Task     Task
	WasError bool
	Output   string
}

func NewScheduler(conf Config) *Scheduler {
	sch := Scheduler{Config: conf}
	sch.TasksQueue = make(chan Task, DefaultChannelSize)
	sch.TasksResultQueue = make(chan TaskResult, DefaultChannelSize)
	sch.NotifiersManager = NewNotifiersManager(sch.TasksResultQueue)
	if conf.Telegram.Enabled {
		sch.NotifiersManager.RegisterNotifier(NewTelegram(conf.Telegram))
	}
	return &sch
}

func (s *Scheduler) Done() {
	close(s.TasksQueue)
	close(s.TasksResultQueue)
}

func (s *Scheduler) Run(ctx context.Context) {
	go s.RunExecutor(ctx)
	go s.NotifiersManager.RunNotifiersManager(ctx)
	for _, task := range s.Config.Tasks {
		s.TasksQueue <- task
	}
	<-ctx.Done()
}

func (s *Scheduler) RunExecutor(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-s.TasksQueue:
			go s.ExecTask(ctx, task)
		}
	}
}

func (s *Scheduler) ExecTask(ctx context.Context, task Task) {
	newCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	cmdWithArgs, _ := shlex.Split(task.Command)
	cmd := exec.CommandContext(newCtx, cmdWithArgs[0], cmdWithArgs[1:]...) //nolint:gosec
	cmd.Dir = task.CWD
	log.Printf("[INFO] Executing task %v", task.ID)

	log.Printf("[DEBUG] Running %v", cmdWithArgs)
	output, err := cmd.CombinedOutput()
	cancel()
	log.Printf("[DEBUG] Result of %v: %v", cmdWithArgs, err)
	strOutput := string(output)
	wasError := false
	if err != nil {
		wasError = true
		log.Printf("[ERROR] Result of %v: %s", cmdWithArgs, err.Error())
		strOutput = fmt.Sprintf("ERROR: %s\nOUTPUT:\n%s", err.Error(), strOutput)
	}
	tr := TaskResult{Task: task, WasError: wasError, Output: strOutput}
	s.TasksResultQueue <- tr
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Duration(task.FrequensySec) * time.Second):
			s.TasksQueue <- task
		}
	}()
}
