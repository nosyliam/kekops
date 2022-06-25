package core

import (
	_ "kekops/ui"
	"strconv"
	//"strings"
	//"time"
)

func BotAll() {
	<-LaunchJobAccounts(func(proxy *Proxy, cookies chan *Account) {
		for {
			var cookie *Account
			select {
			case cookie = <-cookies:
			default:
				return
			}

			if cookie == nil {
				continue
			}
			for _, model := range Tracker.Models {
				proxy.TryRequest(GenerateBotInsert(cookie, strconv.Itoa(model)))
			}

			/*
			for _, model := range Tracker.Models {
				info := <-FetchProductInfo(strconv.Itoa(model))
				proxy.TryRequest(GenerateBotInsert(cookie, strconv.Itoa(model)))
				/*take := proxy.TryRequest(GenerateBotTakeRequestPlugin(cookie.cookie, *info))
				if strings.Contains(take, "Unauthorized") {
					continue
				}
				time.Sleep(50 * time.Millisecond)
				var out string
				if strings.Contains(take, "\"purchased\":true") {
					out = proxy.TryRequest(GenerateBotRemoveRequest(cookie.cookie, strconv.Itoa(model)))
					continue
				}
				if strings.Contains(take, "AlreadyOwned") {
					out = proxy.TryRequest(GenerateBotRemoveRequest(cookie.cookie, strconv.Itoa(model)))
				}
				var tries = 0
				for {
					if !strings.Contains(out, "isValid") && tries < 5  {
						out = proxy.TryRequest(GenerateBotRemoveRequest(cookie.cookie, strconv.Itoa(model)))
						time.Sleep(500 * time.Millisecond)
						tries++
					} else {
						break
					}
				}
				time.Sleep(1000 * time.Millisecond)
			}*/
			cookies <- cookie
		}
	})
}
