package ui

import (
	"fmt"
	ui "kekops/termui/v3"
	"kekops/termui/v3/widgets"
	"time"
)

type Task interface {
	Type() string
	Priority() int
	Destination() string
	Color() string
	Execute()
	Quit()
}

type BaseTask struct {
	Task

	TType               string
	TDestination        interface{}
	TDisplayDestination string

	CompletedRequests int
	Cancel            bool
	CookiesCycled     bool
}

func (b *BaseTask) Type() string {
	return b.TType
}

func (b *BaseTask) Destination() string {
	switch v := b.TDestination.(type) {
	case string:
		return v
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return b.TDisplayDestination
	}
}

func (b *BaseTask) Quit() {
	b.Cancel = true
}

type KekTaskList struct {
	KekUi
	widgets.List

	dragging     bool
	PendingTasks []Task
	CurrentTask  Task
}

func CreateTaskList() *KekTaskList {
	list := &KekTaskList{List: *widgets.NewList()}
	list.Title = "Task List"
	list.SelectedRow = 1
	list.SelectedRowStyle = ui.Style{Bg: ui.ColorBlue}
	list.Rows = []string{"[X] Waiting For Scheduler"}
	go func() {
		n := 0
		ticker := time.NewTicker(time.Millisecond * 50)
		chars := []string{"|", "|", "|", "|", "/", "/", "/", "/", "-", "-", "-", "-", "\\", "\\", "\\", "\\"}
		for {
			<-ticker.C
			if list.CurrentTask != nil {
				list.Rows[0] = "[" + chars[n%len(chars)] + "]" + list.Rows[0][3:]
			}
			ui.Render(list)
			n++
		}
	}()
	return list
}

func (k *KekTaskList) Render() {
	prevRowData := k.Rows[0]
	k.Rows = make([]string, len(k.PendingTasks)+1)
	k.Rows[0] = prevRowData
	if k.CurrentTask == nil {
		k.Rows[0] = "[X] Waiting For Scheduler"
	} else {
		k.Rows[0] = k.Rows[0][0:2] + fmt.Sprintf("  [%s](fg:%s) %10s",
			k.CurrentTask.Type(), k.CurrentTask.Color(), k.CurrentTask.Destination())
	}

	for n, task := range k.PendingTasks {
		k.Rows[n+1] = fmt.Sprintf("[[%d]](fg:green) [%s](fg:%s) %10s",
			n+1, task.Type(), task.Color(), task.Destination())
	}
}

func (k *KekTaskList) HandleEvent(e ui.Event) {
	switch e.ID {
	case "<Up>":
		if k.SelectedRow == 1 {
			return
		}
		k.ScrollUp()
		if k.dragging {
			k.PendingTasks[k.SelectedRow-1], k.PendingTasks[k.SelectedRow] =
				k.PendingTasks[k.SelectedRow], k.PendingTasks[k.SelectedRow-1]
		}
		k.Render()
	case "<Down>":
		if k.SelectedRow == len(k.Rows) {
			return
		}
		k.ScrollDown()
		if k.dragging {
			k.PendingTasks[k.SelectedRow-1], k.PendingTasks[k.SelectedRow-2] =
				k.PendingTasks[k.SelectedRow-2], k.PendingTasks[k.SelectedRow-1]
		}
		k.Render()
	case "<Enter>":
		if k.dragging {
			k.dragging = false
			k.SelectedRowStyle = ui.Style{Bg: ui.ColorBlue}
			return
		}
		k.dragging = true
		k.SelectedRowStyle = ui.Style{Bg: ui.ColorCyan}
	case "g":
	case "<Home>":
		k.ScrollTop()
	case "G", "<End>":
		k.ScrollBottom()
	}
}
