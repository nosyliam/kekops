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

const BotHolder = `_|WARNING:-DO-NOT-SHARE-THIS.--Sharing-this-will-allow-someone-to-log-in-as-you-and-to-steal-your-ROBUX-and-items.|_0B6196759DF2B3E0737172C95F85F798A4E0C5780DEC6D7B3BA6756267ABAF5B6C40A315B8CBB861C00FB6B151FD59777AEED172831B0C1635089E764889A968C65D69F5CB4A97681B4F6CD85AF4C7DF3EF3AECD3AC7E6C9A239E43BF2D284480C96A06AC2145208D2E177D62F84071269E3E623E4723B40068EF5198A95E575E618F8E31444CB5C71CD08C1B0F2E41CB7A95604A5763BC48394CE80E29579A5C04FE636E72CDAFC057664B8DA00325F5AE2252F1CC7BB59045DEE7BFE5BFDF5CA57853994C85C543EC3699D269EEF979E9564D1B6928E15C14EE3AB2DDC31A9065A54FB55C816FC94F95837698E520CE100C2244AF8C4538441F7C5471174EA1909727B47355A981B2DD2C41FA7D7E51EAE9478D5D6261678F7C01FC8DD73B227B08C1EFCCD70AEAF26834FCCD9E26C87E60C275775E4AFC15F6A7CDAD8830B162756DCAAADC6163B9A419379127EC0E5A32BB4`

type AutoPluginFrontPageTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int

	UseHolder bool
}

func (a *AutoPluginFrontPageTask) Execute() {
	imp := a.TDestination.(*Import)
	data := imp.Data()
	if len(data) == 0 {
		return
	}
	var info *ProductInfo
	if imp.Asset != 0 {
		info = <-FetchProductInfo(strconv.Itoa(imp.Asset))
		if info == nil {
			Console.Log(CATEGORY_ROBLOX, LOG_ERROR, "Failed to get product info for FP asset!")
		}
	} else {
		for {
			cookie := BotHolder
			if !a.UseHolder {
				cookie = CookieManager.Verified()
			}
			asset := Proxies[0].TryRequest(GenerateBotUploadPlugin(cookie, data, a.UseHolder))
			if assetId, err := strconv.Atoi(asset); err != nil {
				continue
			} else {
				thumbnail := Proxies[0].TryRequest(GenerateThumbnailRequest(cookie, asset, imp.Icon()))
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Thumbnail upload response: %s", thumbnail))
				info = <-FetchProductInfo(asset)
				if info != nil {
					Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Generating new FP asset ID = %s", asset))
					Tracker.UpdateAsset(imp.ID, assetId)
					/*<-LaunchJobLimited(func(proxy *Proxy, cookies chan string) { // Like
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
							resp := proxy.TryRequest(GenerateBotTakeRequestPlugin(cookie, *info))
							if strings.Contains(resp, "\"purchased\"") {
								proxy.TryRequest(GenerateBotLike(cookie, asset))
							}
							atomic.AddUint64(&a.Count, 1)
							time.Sleep(100 * time.Millisecond)
						}
					}, 1000)*/
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
					}, 1000)
					break
				}
			}
		}
	}
	dest := fmt.Sprintf("%d", info.AssetID)
	a.Goal = uint64(rand.Intn(200000) + 100000) // 1m-2m sales
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
				return
			}
			var cookie string
			select {
			case cookie = <-cookies:
			default:
				return
			}
			take := proxy.TryRequest(GenerateBotTakeRequestPlugin(cookie, *info))
			if take != "" {
				if strings.Contains(take, "\"purchased\":true") {
					atomic.AddUint64(&a.Count, 1)
					proxy.TryRequest(GenerateBotRemoveRequest(cookie, dest))
				}
			}
			cookies <- cookie
			time.Sleep(500 * time.Millisecond)
		}
	})
}

func (a *AutoPluginFrontPageTask) Color() string {
	return "green"
}

func (a *AutoPluginFrontPageTask) Type() string {
	return "AUTOFP"
}
