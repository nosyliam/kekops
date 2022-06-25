package core

import (
	"fmt"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	"strings"

	//	"strings"
	"sync/atomic"
	"time"
)

type BotTask struct {
	BaseTask

	Goal  uint64
	Count uint64
}

func (b *BotTask) Execute() {
	dest := fmt.Sprintf("%d", b.TDestination)
	info := <-FetchProductInfo(dest)

	if info == nil {
		Console.Log(CATEGORY_ROBLOX, LOG_WARN, "Failed to fetch product info; skipping task")
		return
	}
	go func() {
		for {
			if b.Count >= b.Goal || b.Cancel == true {
				b.Cancel = true
				break
			}
			Gauge.Percent = ui.MinInt(int(math.Ceil((float64(atomic.LoadUint64(&b.Count))/float64(b.Goal))*100)), 100)
			ui.Render(Gauge)
			time.Sleep(10 * time.Millisecond)
		}
	}()
	<-LaunchJob(func(proxy *Proxy, cookies chan string) {
		for {
			if b.Cancel {
				return
			}
			var cookie string
			select {
			case cookie = <-cookies:
			default:
				return
			}
			var take string
			switch info.AssetTypeID {
			case 38: // Plugin
				take = proxy.TryRequest(GenerateBotTakeRequestPlugin(cookie, *info))
			case 10: // Model
				take = proxy.TryRequest(GenerateBotTakeRequestPlugin(cookie, *info))
			}
			/*if strings.Contains(take, "TooManyRequests") || strings.Contains(take, "TooManyPurchases") {
				time.Sleep(1 * time.Second)
				cookies <- cookie
				continue
			}*/
			if strings.Contains(take, "Unauthorized") {
				continue
			}
			time.Sleep(50 * time.Millisecond)
			var out string
			if strings.Contains(take, "\"purchased\":true") {
				out = proxy.TryRequest(GenerateBotRemoveRequest(cookie, dest))
				atomic.AddUint64(&b.Count, 1)
			}
			if strings.Contains(take, "AlreadyOwned") {
				out = proxy.TryRequest(GenerateBotRemoveRequest(cookie, dest))
			}
			var tries = 0
			for {
				if !strings.Contains(out, "isValid") && tries < 5  {
					out = proxy.TryRequest(GenerateBotRemoveRequest(cookie, dest))
					time.Sleep(500 * time.Millisecond)
					tries++
				} else {
					break
				}
			}
			time.Sleep(1000 * time.Millisecond)
			cookies <- cookie
		}
	})
}

func (b *BotTask) Color() string {
	return "yellow"
}

func (b *BotTask) Type() string {
	return "BOT"
}
