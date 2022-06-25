package ui

import (
	"fmt"
	"image"
	ui "kekops/termui/v3"
	"kekops/termui/v3/widgets"
	"time"
)

const (
	xAxisLabelsHeight = 0
	yAxisLabelsWidth  = 6
	xAxisLabelsGap    = 2
	yAxisLabelsGap    = 1
)

type KekChart struct {
	widgets.Plot
	KekUi

	requests float64
}

func CreateChart() *KekChart {
	chart := &KekChart{Plot: *widgets.NewPlot()}
	chart.Title = "Requests Per Second"
	chart.Marker = widgets.MarkerDot
	chart.Data = [][]float64{[]float64{0}}
	chart.AxesColor = ui.ColorWhite
	chart.LineColors[0] = ui.ColorYellow
	chart.DrawDirection = widgets.DrawLeft
	go func() {
		ticker := time.NewTicker(time.Second * 1)
		for {
			<-ticker.C
			chart.Data[0] = append(chart.Data[0], chart.requests)
			if len(chart.Data[0]) == 55 {
				chart.Data[0] = chart.Data[0][1:54]
			}
			chart.requests = 0
			ui.Render(chart)
		}
	}()
	return chart
}

func (k *KekChart) Increment() {
	k.requests += 1
}

func (self *KekChart) plotAxes(buf *ui.Buffer, maxVal float64) {
	buf.SetCell(
		ui.NewCell(ui.BOTTOM_LEFT, ui.NewStyle(ui.ColorWhite)),
		image.Pt(self.Inner.Min.X+yAxisLabelsWidth, self.Inner.Max.Y-xAxisLabelsHeight-1),
	)
	// draw x axis line
	for i := yAxisLabelsWidth + 1; i < self.Inner.Dx(); i++ {
		buf.SetCell(
			ui.NewCell(ui.HORIZONTAL_DASH, ui.NewStyle(ui.ColorWhite)),
			image.Pt(i+self.Inner.Min.X, self.Inner.Max.Y-xAxisLabelsHeight-1),
		)
	}
	// draw y axis line
	for i := 0; i < self.Inner.Dy()-xAxisLabelsHeight-1; i++ {
		buf.SetCell(
			ui.NewCell(ui.VERTICAL_LINE, ui.NewStyle(ui.ColorWhite)),
			image.Pt(self.Inner.Min.X+yAxisLabelsWidth, i+self.Inner.Min.Y),
		)
	}
	// draw y axis labels
	verticalScale := maxVal / float64(self.Inner.Dy()-xAxisLabelsHeight-1)
	for i := 0; i*(yAxisLabelsGap+1) < self.Inner.Dy()-1; i++ {
		buf.SetString(
			fmt.Sprintf("%d", int(float64(i)*verticalScale*(yAxisLabelsGap+1))),
			ui.NewStyle(ui.ColorWhite),
			image.Pt(self.Inner.Min.X, self.Inner.Max.Y-(i*(yAxisLabelsGap+1))-2),
		)
	}
}

func (self *KekChart) renderDot(buf *ui.Buffer, drawArea image.Rectangle, maxVal float64) {
	for i, line := range self.Data {
		for j := 0; j < len(line) && j*self.HorizontalScale < drawArea.Dx(); j++ {
			val := line[j]
			height := int((val / maxVal) * float64(drawArea.Dy()-1))
			buf.SetCell(
				ui.NewCell(self.DotMarkerRune, ui.NewStyle(ui.SelectColor(self.LineColors, i))),
				image.Pt(drawArea.Min.X+(j*self.HorizontalScale), drawArea.Max.Y-1-height),
			)
		}
	}
}

func (self *KekChart) Draw(buf *ui.Buffer) {
	self.Block.Draw(buf)

	maxVal := self.MaxVal
	if maxVal == 0 {
		maxVal, _ = ui.GetMaxFloat64From2dSlice(self.Data)
	}

	if self.ShowAxes {
		self.plotAxes(buf, maxVal)
	}

	drawArea := image.Rect(
		self.Inner.Min.X+yAxisLabelsWidth+1, self.Inner.Min.Y+1,
		self.Inner.Max.X, self.Inner.Max.Y-xAxisLabelsHeight-1,
	)

	self.renderDot(buf, drawArea, maxVal)
}
