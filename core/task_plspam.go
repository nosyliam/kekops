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

type SpamPluginTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int
}

func (s *SpamPluginTask) Execute() {
	imp := s.TDestination.(*Import)
	data := imp.Data()
	if len(data) == 0 {
		return
	}
	s.Goal = 50
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
		upload := proxy.TryRequest(GenerateBotUploadPluginSpam(BotHolder, imp.Name, data, 100, 1000, ""))
		if upload != "" {
			Console.Log(CATEGORY_ROBLOX, LOG_INFO, upload)
		}
		if _, err := strconv.Atoi(upload); err != nil {
			return
		}
		for {
			resp := proxy.TryRequest(GenerateThumbnailRequest(BotHolder, upload, imp.Icon()))
			if strings.Contains(resp, "target") {
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Uploaded icon: %s", resp))
				break
			} else {
				Console.Log(CATEGORY_ROBLOX, LOG_TRACE, fmt.Sprintf("Failed icon: %s", resp))
			}
			time.Sleep(5 * time.Second)
		}
		info := <-FetchProductInfo(upload)
		if info == nil {
			return
		}
		<-LaunchJobLimited(func(proxy *Proxy, cookies chan string) { // Like
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
				proxy.TryRequest(GenerateBotTakeRequest(cookie, *info))
				time.Sleep(100 * time.Millisecond)
			}
		}, int(rand.Int31n(100)))
		atomic.AddUint64(&s.Count, 1)
	})
}

func (b *SpamPluginTask) Color() string {
	return "green"
}

func (b *SpamPluginTask) Type() string {
	return "SPAM"
}
