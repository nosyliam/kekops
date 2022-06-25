package core

import (
	"fmt"
	math2 "github.com/google/gxui/math"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

type DislikeTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int
}

func (l *DislikeTask) Execute() {
	dest := fmt.Sprintf("%d", l.TDestination)
	l.Goal = uint64(len(CookieManager.cookies))
	info := <-FetchProductInfo(dest)
	go func() {
		for {
			if l.Count >= l.Goal || l.Cancel == true {
				l.Cancel = true
				break
			}
			Gauge.Percent = math2.Min(int(math.Ceil((float64(atomic.LoadUint64(&l.Count))/float64(l.Goal))*100)), 100)
			ui.Render(Gauge)
			time.Sleep(10 * time.Millisecond)
		}
	}()
	<-LaunchJob(func(proxy *Proxy, cookies chan string) {
		for {
			if len(cookies) == 0 {
				return
			}
			cookie := <-cookies
			if l.Cancel == true {
				return
			}
			switch info.AssetTypeID {
			case 38: // Plugin
				proxy.TryRequest(GenerateBotRemoveRequest(cookie, dest))
				resp := proxy.TryRequest(GenerateBotTakeRequestPlugin(cookie, *info))
				if resp == "403" {
					continue
				}
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, resp)
				if strings.Contains(resp, "\"purchased\""){
					proxy.TryRequest(GenerateBotDislike(cookie, dest))
				}
			case 10: // Model
				//resp := proxy.TryRequest(GenerateBotInsert(cookie, dest))
				//if resp == "true" {
				//	proxy.TryRequest(GenerateBotDislike(cookie, dest))
				//}
			}
			atomic.AddUint64(&l.Count, 1)
			time.Sleep(100 * time.Millisecond)
		}
	})
}

func (l *DislikeTask) Color() string {
	return "yellow"
}

func (b *DislikeTask) Type() string {
	return "DSLK"
}
