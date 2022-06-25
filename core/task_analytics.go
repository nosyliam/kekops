package core

import (
	"fmt"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	"sync/atomic"
	"time"
)

type AnalyticsTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int

	QueryName string
}

func (a *AnalyticsTask) Execute() {
	dest := fmt.Sprintf("%d", a.TDestination)
	Console.Log(CATEGORY_ROBLOX, LOG_INFO, a.QueryName)
	go func() {
		for {
			if a.Count >= a.Goal || a.Cancel == true {
				a.Cancel = true
				break
			}
			Gauge.Percent = ui.MinInt(int(math.Ceil((float64(atomic.LoadUint64(&a.Count))/float64(a.Goal))*100)), 100)
			ui.Render(Gauge)
			time.Sleep(10 * time.Millisecond)
		}
	}()
	<-LaunchJob(func(proxy *Proxy, cookies chan string) {
		for {
			if a.Cancel {
				return
			}
			var cookie string
			select {
			case cookie = <-cookies:
			default:
				return
			}
			proxy.TryRequest(GenerateAnalytics1(cookie, dest))
			proxy.TryRequest(GenerateAnalytics2(cookie, dest, a.QueryName))
			proxy.TryRequest(GenerateAnalytics3(cookie, dest))
			atomic.AddUint64(&a.Count, 1)
			cookies <- cookie
			time.Sleep(500 * time.Millisecond)
		}
	})
}

func (a *AnalyticsTask) Color() string {
	return "green"
}

func (a *AnalyticsTask) Type() string {
	return "ECSV2"
}
