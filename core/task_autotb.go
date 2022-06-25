package core

import (
	"fmt"
	math2 "github.com/google/gxui/math"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type AutoToolboxTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int
}

func (a *AutoToolboxTask) Execute() {
	defer func() {
		if r := recover(); r != nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Caught panic!")
		}
	}()

	// Upload five new models on verified accounts, spam insert, like/fav bot them
	imp := a.TDestination.(*Import)
	data := imp.Data()
	if len(data) == 0 {
		return
	}
	Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Executing auto-toolbox operation on import %s", imp.Name))
	go func() {
		for {
			if a.Goal <= a.Count || a.Cancel == true {
				a.Cancel = true
				return
			}
			Gauge.Percent = math2.Min(int(math.Ceil((float64(atomic.LoadUint64(&a.Count))/float64(a.Goal))*100)), 100)
			ui.Render(Gauge)
			time.Sleep(10 * time.Millisecond)
		}
	}()
	// Do a little spamming first
	a.Goal = 1000
	<-LaunchJob(func(proxy *Proxy, cookies chan string) {
		for {
			if a.Cancel {
				return
			}
			cookie := <-cookies
			upload := proxy.TryRequest(GenerateBotUpload(cookie, imp.Name, data, true, false))
			cookies <- cookie
			if upload != "" {
				Console.Log(CATEGORY_ROBLOX, LOG_TRACE, upload)
				atomic.AddUint64(&a.Count, 1)
			}
			if _, err := strconv.Atoi(upload); err != nil {
				continue
			}
			info := <-FetchProductInfo(upload)
			if info == nil {
				continue
			}
			for n := 0; n < 5; n++ {
				if a.Cancel {
					return
				}
				cookie := <-cookies
				proxy.TryRequest(GenerateBotTakeRequest(cookie, *info))
				cookies <- cookie
				time.Sleep(200 * time.Millisecond)
			}
		}
	})
	time.Sleep(500 * time.Millisecond)
	go func() {
		for {
			if a.Cancel == true {
				return
			}
			Gauge.Percent = math2.Min(int(math.Ceil((float64(atomic.LoadUint64(&a.Count))/float64(a.Goal))*100)), 100)
			ui.Render(Gauge)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	var uploaded []string
	for n := 0; n < 3; {
		cookie := CookieManager.Verified()
		upload := Proxies[0].TryRequest(GenerateBotUpload(cookie, imp.ToolboxName, data, false, false))
		if _, err := strconv.Atoi(upload); err != nil {
			continue
		}
		uploaded = append(uploaded, upload)
		n++
	}
	Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Uploaded [%d](fg:green) assets to toolbox bot", len(uploaded)))
	for _, asset := range uploaded {
		a.Goal += uint64(len(CookieManager.cookies))
		info := <-FetchProductInfo(asset)
		if info == nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_WARN, fmt.Sprintf("Failed to get product info for auto-toolbox asset %s", asset))
			continue
		}
		<-LaunchJob(func(proxy *Proxy, cookies chan string) {
			for n := 0; n < 5; n++ {
				cookie := <-cookies
				//proxy.TryRequest(GenerateBotInsert(cookie, asset))
				cookies <- cookie
			}
		})
		<-LaunchJob(func(proxy *Proxy, cookies chan string) { // Like
			for {
				if len(cookies) == 0 {
					return
				}
				cookie := <-cookies
				switch info.AssetTypeID {
				case 38: // Plugin
					resp := proxy.TryRequest(GenerateBotTakeRequest(cookie, *info))
					if strings.Contains(resp, "\"purchased\"") {
						//proxy.TryRequest(GenerateBotLike(cookie, asset))
					}
				case 10: // Model
					//resp := proxy.TryRequest(GenerateBotInsert(cookie, asset))
					//if resp == "true" {
						//proxy.TryRequest(GenerateBotLike(cookie, asset))
					//}
				}
				atomic.AddUint64(&a.Count, 1)
				time.Sleep(100 * time.Millisecond)
			}
		})
		a.Goal += uint64(len(CookieManager.cookies))
		<-LaunchJob(func(proxy *Proxy, cookies chan string) { // Favorite
			for {
				if len(cookies) == 0 {
					return
				}
				cookie := <-cookies
				proxy.TryRequest(GenerateBotFavorite(cookie, asset))
				atomic.AddUint64(&a.Count, 1)
				time.Sleep(100 * time.Millisecond)
			}
		})
	}
	Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf(
		"Finished automatic operations: assets = %s", strings.Join(uploaded, ", ")))

}

func (a *AutoToolboxTask) Color() string {
	return "green"
}

func (a *AutoToolboxTask) Type() string {
	return "AUTOTB"
}
