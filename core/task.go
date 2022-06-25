package core

import (
	"fmt"
	. "kekops/ui"
)

var Scheduler *KekScheduler

type KekScheduler struct {
	Queue chan Task
}

func (k *KekScheduler) ExecuteNext() {
	if len(TaskList.PendingTasks) == 0 || TaskList.CurrentTask != nil {
		TaskList.Render()
		return
	}
	TaskList.CurrentTask = TaskList.PendingTasks[0]
	TaskList.PendingTasks = TaskList.PendingTasks[1:]
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Captured panic: %v", r))
			}
		}()
		TaskList.CurrentTask.Execute()
		TaskList.CurrentTask = nil
		Gauge.Percent = 0
		k.ExecuteNext()
	}()
	TaskList.Render()
}

func (k *KekScheduler) Poll() {
	for {
		defer func() {
			if r := recover(); r != nil {
				Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Captured panic: %v", r))
			}
		}()

		task := <-k.Queue
		TaskList.PendingTasks = append(TaskList.PendingTasks, task)
		TaskList.Render()
		if TaskList.CurrentTask == nil {
			k.ExecuteNext()
		}
	}
}

func init() {
	Scheduler = &KekScheduler{Queue: make(chan Task)}
}
