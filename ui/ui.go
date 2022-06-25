package ui

import (
	"log"

	tb "github.com/gdamore/tcell/termbox"
	ui "kekops/termui/v3"
	"kekops/termui/v3/widgets"
)

var (
	Input      *widgets.Paragraph
	Console    *KekConsole
	TaskList   *KekTaskList
	ErrorChart *KekErrorChart
	Gauge      *KekGauge
	Chart      *KekChart

	Focused KekUi
)

type KekUi interface {
	HandleEvent(event ui.Event)
}

func InitUi() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Console + input
	Console = NewConsole()
	Console.SetRect(0, 0, 50, 31)

	Input = widgets.NewParagraph()
	Input.SetRect(0, 31, 50, 35)
	Input.Text = "> "

	TaskList = CreateTaskList()
	TaskList.SetRect(52, 0, 82, 10)

	ErrorChart = CreateErrorChart()
	ErrorChart.SetRect(84, 0, 116, 10)

	Gauge = CreateGauge()
	Gauge.SetRect(52, 11, 116, 14)

	Chart = CreateChart()
	Chart.SetRect(52, 15, 116, 35)

	ui.Render(Console, Input, TaskList, ErrorChart, Gauge, Chart)
	tb.SetCursor(3, 32)

	Focused = Console

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<F1>":
				Focused = Console
			case "<F2>":
				Focused = TaskList
			default:
				Focused.HandleEvent(e)
			}
		case <-ExitChan:
			return
		}
	}
}
