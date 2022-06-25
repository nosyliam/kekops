package core

import (
	"bytes"
	"fmt"
	"io/ioutil"
	. "kekops/ui"
	"net/http"
	"net/url"
	"strings"
	"time"
	//"time"
)

var Proxies = []*Proxy{&Proxy{client: &http.Client{}}}

type Proxy struct {
	client *http.Client
	prev   *Proxy
}

func (p *Proxy) TryRequest(req *http.Request) string {
	defer func() {
		if r := recover(); r != nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Caught panic!")
		}
	}()

	var resp []byte
	reqData, _ := ioutil.ReadAll(req.Body)
	try := func(px *Proxy) <-chan int {
		out := make(chan int)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Caught panic!")
				}
			}()
			defer req.Body.Close()
			req.Body = ioutil.NopCloser(bytes.NewBuffer(reqData))
			pResp, err := px.client.Do(req)
			if err != nil {
				Console.Log(CATEGORY_ROBLOX, LOG_TRACE, fmt.Sprintf("%v", err))
				ErrorChart.IncrementHTTP()
				out <- 0
				return
			}
			Chart.Increment()
			defer pResp.Body.Close()
			resp, _ = ioutil.ReadAll(pResp.Body)
			if strings.Contains(string(resp), "TooManyRequests") {
				ErrorChart.IncrementAPI()
				out <- 3
				return
			}
			Console.Log(CATEGORY_ROBLOX, LOG_TRACE, fmt.Sprintf("%d", pResp.StatusCode))
			switch pResp.StatusCode {
			case 400:
			case 401:
				ErrorChart.Increment400()
				out <- 0
				return
			case 403:
				ErrorChart.Increment403()
				out <- 1
				return
			case 500:
				ErrorChart.Increment500()
				out <- 0
				return
			}
			out <- 2
		}()
		return out
	}
	var proxy = p
	for n := 0; n < 3; n++ {
		after := time.After(15 * time.Second)
		select {
		case <-after:
			Console.Log(CATEGORY_ROBLOX, LOG_TRACE, "Request timed out! Trying another proxy...")
			//proxy = proxy.prev
			return ""
		case status := <-try(proxy):
			switch status {
			case 1:
				return "403"
			case 2:
				return string(resp)
			case 3:
				time.Sleep(5 * time.Second)
			}
		}
	}
	return ""
}

func LoadProxies(file string) {
	for {
		if Console != nil {
			break
		}
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Failed to open %s: %v", file, err))
		return
	}
	proxies := strings.SplitN(string(data), "\n", -1)
	var prev = Proxies[0]
	for _, proxy := range proxies {
		split := strings.SplitN(strings.TrimSpace(proxy), ":", -1)
		var url *url.URL
		switch len(split) {
		case 2: // IP authentication
			url, _ = url.Parse("http://" + split[0] + ":" + split[1])
		case 4: // Username & password
			url, _ = url.Parse("http://mpaboxvl-dest:unoso0hwve7w@" + split[0] + ":" + split[1])
		default:
			continue
		}
		Proxies = append(Proxies, &Proxy{client: &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(url),
			},
		}, prev: prev})
	}
	Console.Log(CATEGORY_MAIN, LOG_INFO, fmt.Sprintf("Loaded %d proxies.", len(proxies)))
}

func init() {
	Proxies[0].prev = Proxies[0]
}
