package ui

import (
	ui "kekops/termui/v3"
	"kekops/termui/v3/widgets"
)

type KekGauge struct {
	KekUi
	widgets.Gauge
}

func CreateGauge() *KekGauge {
	gauge := &KekGauge{Gauge: *widgets.NewGauge()}
	gauge.Title = "Task Progress"
	gauge.Percent = 0
	gauge.BarColor = ui.ColorGreen
	gauge.LabelStyle = ui.NewStyle(ui.ColorYellow)
	gauge.TitleStyle.Fg = ui.ColorMagenta
	gauge.BorderStyle.Fg = ui.ColorWhite
	return gauge
}
