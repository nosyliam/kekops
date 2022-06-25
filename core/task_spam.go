package core

import (
	math2 "github.com/google/gxui/math"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	"strconv"
	"sync/atomic"
	"time"
)

type SpamTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int
}

func (s *SpamTask) Execute() {
	imp := s.TDestination.(*Import)
	data := imp.Data()
	if len(data) == 0 {
		return
	}
	s.Goal = 1000
	go func() {
		for {
			if s.Count >= s.Goal || s.Cancel == true {
				s.Cancel = true
				break
			}
			Gauge.Percent = math2.Min(int(math.Ceil((float64(atomic.LoadUint64(&s.Count))/float64(s.Goal))*100)), 100)
			ui.Render(Gauge)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	<-LaunchJob(func(proxy *Proxy, cookies chan string) {
		for {
			if s.Cancel {
				return
			}
			cookie := <-cookies
			upload := proxy.TryRequest(GenerateBotUpload(cookie, imp.Name, data, true, true))
			cookies <- cookie
			if upload != "" {
				Console.Log(CATEGORY_ROBLOX, LOG_TRACE, upload)
			}
			if _, err := strconv.Atoi(upload); err != nil {
				continue
			}
			info := <-FetchProductInfo(upload)
			if info == nil {
				continue
			}
			for n := 0; n < 50; n++ {
				if s.Cancel {
					return
				}
				cookie := <-cookies
				proxy.TryRequest(GenerateBotTakeRequest(cookie, *info))
				proxy.TryRequest(GenerateAnalytics1(cookie, upload))
				proxy.TryRequest(GenerateAnalytics2(cookie, upload, imp.Name))
				proxy.TryRequest(GenerateAnalytics3(cookie, upload))
				cookies <- cookie
				time.Sleep(200 * time.Millisecond)
			}
			atomic.AddUint64(&s.Count, 1)
		}
	})
}

func (b *SpamTask) Color() string {
	return "green"
}

func (b *SpamTask) Type() string {
	return "SPAM"
}
