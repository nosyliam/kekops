package ui

import (
	rw "github.com/mattn/go-runewidth"
	"image"
	ui "kekops/termui/v3"
	"kekops/termui/v3/widgets"
	"time"
)

type KekErrorChart struct {
	KekUi
	widgets.BarChart
}

func CreateErrorChart() *KekErrorChart {
	chart := &KekErrorChart{BarChart: *widgets.NewBarChart()}
	chart.Data = []float64{0, 0, 0, 0, 0}
	chart.Labels = []string{"400", "403", "500", "HTTP", "API"}
	chart.Title = "Error Rates"
	chart.BarWidth = 3
	chart.BarGap = 3
	chart.BarColors = []ui.Color{ui.ColorRed}
	chart.LabelStyles = []ui.Style{ui.NewStyle(ui.ColorWhite)}
	chart.NumStyles = []ui.Style{ui.NewStyle(ui.ColorYellow)}
	go func() {
		ticker := time.NewTicker(time.Second * 2)
		for {
			<-ticker.C
			ui.Render(chart)
			chart.Data = []float64{0, 0, 0, 0, 0}
		}
	}()
	return chart
}

func (k *KekErrorChart) Increment400() {
	k.Data[0] += 1
}

func (k *KekErrorChart) Increment403() {
	k.Data[1] += 1
}

func (k *KekErrorChart) Increment500() {
	k.Data[2] += 1
}

func (k *KekErrorChart) IncrementHTTP() {
	k.Data[3] += 1
}

func (k *KekErrorChart) IncrementAPI() {
	k.Data[4] += 1
}

// Re-implementation of bar chart because Zack Guo can't do it right
func (self *KekErrorChart) Draw(buf *ui.Buffer) {
	self.Block.Draw(buf)

	maxVal := self.MaxVal
	if maxVal == 0 {
		maxVal, _ = ui.GetMaxFloat64FromSlice(self.Data)
	}

	barXCoordinate := self.Inner.Min.X + 1

	for i, data := range self.Data {
		if data > 0 {
			height := ui.MaxInt(1, int((data/maxVal)*float64(self.Inner.Dy()-2)))
			for x := barXCoordinate; x < ui.MinInt(barXCoordinate+self.BarWidth, self.Inner.Max.X); x++ {
				for y := self.Inner.Max.Y - 2; y > (self.Inner.Max.Y-2)-height; y-- {
					c := ui.NewCell(' ', ui.NewStyle(ui.ColorClear, ui.SelectColor(self.BarColors, i)))
					buf.SetCell(c, image.Pt(x, y))
				}
			}
		}

		if i < len(self.Labels) {
			labelXCoordinate := barXCoordinate +
				int((float64(self.BarWidth) / 2)) -
				int((float64(rw.StringWidth(self.Labels[i])) / 2))
			buf.SetString(
				self.Labels[i],
				ui.SelectStyle(self.LabelStyles, i),
				image.Pt(labelXCoordinate, self.Inner.Max.Y-1),
			)
		}

		numberXCoordinate := barXCoordinate
		if numberXCoordinate <= self.Inner.Max.X {
			var style ui.Style
			if data <= 0 {
				style = ui.NewStyle(ui.SelectStyle(self.NumStyles, i+1).Fg)
			} else {
				style = ui.NewStyle(
					ui.SelectStyle(self.NumStyles, i+1).Fg,
					ui.SelectColor(self.BarColors, i),
					ui.SelectStyle(self.NumStyles, i+1).Modifier,
				)
			}
			buf.SetString(
				self.NumFormatter(data),
				style,
				image.Pt(numberXCoordinate, self.Inner.Max.Y-2),
			)
		}

		barXCoordinate += (self.BarWidth + self.BarGap)
	}
}
