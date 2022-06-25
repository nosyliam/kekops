package core

import (
	"fmt"
	math2 "github.com/google/gxui/math"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	"sync/atomic"
	"time"
)

type FavoriteTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int
}

func (f *FavoriteTask) Execute() {
	dest := fmt.Sprintf("%d", f.TDestination)
	f.Goal = uint64(len(CookieManager.cookies))
	go func() {
		for {
			if f.Count >= f.Goal || f.Cancel == true {
				f.Cancel = true
				break
			}
			Gauge.Percent = math2.Min(int(math.Ceil((float64(atomic.LoadUint64(&f.Count))/float64(f.Goal))*100)), 100)
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
			if f.Cancel == true {
				return
			}
			proxy.TryRequest(GenerateBotFavorite(cookie, dest))
			atomic.AddUint64(&f.Count, 1)
			cookies <- cookie
			time.Sleep(100 * time.Millisecond)
		}
	})
}

func (f *FavoriteTask) Color() string {
	return "white"
}

func (f *FavoriteTask) Type() string {
	return "FAVE"
}
