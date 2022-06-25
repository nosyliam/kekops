package core

import (
	"fmt"
	math2 "github.com/google/gxui/math"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type AutoFrontPageTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int

	UseHolder bool
}

func (a *AutoFrontPageTask) Execute() {
	imp := a.TDestination.(*Import)
	data := imp.Data()
	if len(data) == 0 {
		return
	}
	a.Count = 1
	var info *ProductInfo
	if imp.Asset != 0 {
		info = <-FetchProductInfo(strconv.Itoa(imp.Asset))
		if info == nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Failed to get product info for FP asset!")
		}
	} else {
		for {
			cookie := BotHolder
			if !a.UseHolder {
				cookie = CookieManager.Verified()
			}
			asset := Proxies[0].TryRequest(GenerateBotUpload(cookie, imp.ToolboxName, data, false, a.UseHolder))
			if assetId, err := strconv.Atoi(asset); err != nil {
				continue
			} else {
				info = <-FetchProductInfo(asset)
				if info != nil {
					Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Generating new FP asset ID = %s", asset))
					Tracker.UpdateAsset(imp.ID, assetId)
					<-LaunchJob(func(proxy *Proxy, cookies chan string) { // Like
						for {
							if len(cookies) == 0 {
								return
							}
							var cookie string
							select {
							case cookie = <-cookies:
							default:
								return
							}
							switch info.AssetTypeID {
							case 38: // Plugin
								resp := proxy.TryRequest(GenerateBotTakeRequest(cookie, *info))
								if strings.Contains(resp, "\"purchased\"") {
									//proxy.TryRequest(GenerateBotLike(cookie, asset))
								}
							case 10: // Model

							}
							atomic.AddUint64(&a.Count, 1)
							time.Sleep(100 * time.Millisecond)
						}
					})
					a.Goal += uint64(len(CookieManager.cookies))
					<-LaunchJobLimited(func(proxy *Proxy, cookies chan string) { // Favorite
						for {
							if len(cookies) == 0 {
								return
							}
							cookie := <-cookies
							proxy.TryRequest(GenerateBotFavorite(cookie, asset))
							atomic.AddUint64(&a.Count, 1)
							time.Sleep(100 * time.Millisecond)
						}
					}, rand.Intn(200)+300)
					break
				}
			}
		}
	}
	dest := fmt.Sprintf("%d", info.AssetID)
	a.Goal = uint64(rand.Intn(200000) + 500000) // 1m-2m sales
	go func() {
		for {
			if a.Count >= a.Goal || a.Cancel == true {
				a.Cancel = true
				break
			}
			Gauge.Percent = math2.Min(int(math.Ceil((float64(atomic.LoadUint64(&a.Count))/float64(a.Goal))*100)), 100)
			ui.Render(Gauge)
			time.Sleep(10 * time.Millisecond)
		}
	}()
	a.Count = 0
	Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Botting for [%d](fg:yellow) sales", a.Goal))
	<-LaunchJob(func(proxy *Proxy, cookies chan string) {
		for {
			if a.Cancel {
				break
			}
			if len(cookies) == 0 {
				return
			}
			var cookie string
			select {
			case cookie = <-cookies:
			default:
				break
			}
			proxy.TryRequest(GenerateBotRemoveRequest(cookie, dest))
			take := proxy.TryRequest(GenerateBotTakeRequestPlugin(cookie, *info))
			if take != "" {
				if strings.Contains(take, "\"purchased\":true") {
					atomic.AddUint64(&a.Count, 1)
				}
			}
			cookies <- cookie
			time.Sleep(500 * time.Millisecond)
		}
	})
}

func (a *AutoFrontPageTask) Color() string {
	return "green"
}

func (a *AutoFrontPageTask) Type() string {
	return "AUTOFP"
}
