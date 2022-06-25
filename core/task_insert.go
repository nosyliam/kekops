package core

import (
	"fmt"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	//	"strings"
	"sync/atomic"
	"time"
)

type InsertTask struct {
	BaseTask

	Goal  uint64
	Count uint64
}

func (b *InsertTask) Execute() {
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
	<-LaunchJobAccounts(func(proxy *Proxy, cookies chan *Account) {
		for {
			if b.Cancel {
				return
			}
			var cookie *Account
			select {
			case cookie = <-cookies:
			default:
				return
			}

			if cookie == nil {
				continue
			}

			resp := proxy.TryRequest(GenerateBotInsert(cookie, dest))
			atomic.AddUint64(&b.Count, 1)
			Console.Log(CATEGORY_ROBLOX, LOG_TRACE, resp)
			time.Sleep(500 * time.Millisecond)
			cookies <- cookie
		}
	})
}

func (b *InsertTask) Color() string {
	return "yellow"
}

func (b *InsertTask) Type() string {
	return "INSERT"
}
