package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	api2captcha "github.com/2captcha/2captcha-go"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	. "kekops/ui"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

func addHeaders(r *http.Request) {
	r.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	r.Header.Add("Upgrade-Insecure-Requests", "1")
	r.Header.Add("Accept-Language", "en-US,en;q=0.9")
	r.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36")
}

type LoginRequest struct {
	CaptchaProvider string `json:"captchaProvider"`
	CaptchaToken string `json:"captchaToken"`
	Ctype        string `json:"ctype"`
	Cvalue       string `json:"cvalue"`
	Cpassword    string `json:"password"`
}

func Login(account *Account) <-chan bool {
	finishChan := make(chan bool)
	go func() {
		Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Starting login for %s", account.user))
		jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		if err != nil {
			finishChan <- false
			return
		}

		u, _ := url.Parse("https://www.roblox.com")
		//proxyUrl, err := url.Parse("http://customer-amcd19228101-asn-20057-sessid-cOWU-sesstime-30:CMTJbgJ5b@residential.ipb.cloud:7777")
		client := Proxies[rand.Intn(len(Proxies))].client
		client.Jar = jar
		//&http.Client{
		//	Jar: jar,
		//	Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
		//}

		req, err := http.NewRequest("GET", "https://www.roblox.com/", nil)
		addHeaders(req)
		client.Do(req)
		req, err = http.NewRequest("GET", "https://www.roblox.com/timg/rbx", nil)
		addHeaders(req)
		client.Do(req)
		if err != nil {
			Console.Log(CATEGORY_ROBLOX, LOG_ERROR, fmt.Sprintf("Unable to load headers: %s", err))

		}

		captcha := api2captcha.NewClient("de731db23a13e12db739d2d90f0dcb61")
		captcha.DefaultTimeout = 100
		captcha.PollingInterval = 5
		cap := api2captcha.FunCaptcha{
			SiteKey:   "476068BF-9607-4799-B53D-966BE98E2B81",
			Url:       "https://www.roblox.com/",
			Surl:      "https://roblox-api.arkoselabs.com",
			UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.61 Safari/537.36",
		}
		creq := cap.ToRequest()
		creq.SetProxy("HTTP", "customer-amcd19228101-cc-GB-sessid-oGCm-sesstime-1:CMTJbgJ5b@residential.ipb.cloud:7777")
		code, err := captcha.Solve(creq)
		if err != nil {
			finishChan <- false
			return
		}
		Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("%s, %s", code, err))

		login := &LoginRequest{
			CaptchaProvider: "PROVIDER_ARKOSE_LABS",
			CaptchaToken:    code,
			Ctype:           "Username",
			Cvalue:          account.user,
			Cpassword:       account.pass,
		}
		data, _ := json.Marshal(login)
		Console.Log(CATEGORY_ROBLOX, LOG_TRACE, fmt.Sprintf("%s", data))
		tryRequest:
		req, _ = http.NewRequest("POST", "https://auth.roblox.com/v2/login", bytes.NewReader(data))
		addHeaders(req)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		resp, err := client.Do(req)
		if err != nil {
			Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("%s", err.Error()))
			goto tryRequest
		}
		xsrf := resp.Header.Get("x-csrf-token")
		req, _ = http.NewRequest("POST", "https://auth.roblox.com/v2/login", bytes.NewReader(data))
		addHeaders(req)
		Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("XSRF: %s", xsrf))
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		req.Header.Add("X-CSRF-TOKEN", xsrf)
		resp, err = client.Do(req)
		if err != nil {
			Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("%s", err.Error()))
			goto tryRequest
		}
		defer resp.Body.Close()

		rdata, _ := ioutil.ReadAll(resp.Body)
		Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("%s", rdata, resp.Header.Get("set-cookie")))
		if strings.Contains(string(rdata), "robot") {
			finishChan <- false
			return
		}
		for _, cookie := range jar.Cookies(u) {
			if cookie.Name == ".ROBLOSECURITY" {
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Received cookie for %s", account.user))
				account.cookie = cookie.Value
			}
			if cookie.Name == ".RBXID" {
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Received ID for %s", account.user))
				account.id = cookie.Value
			}
		}
		finishChan <- true
	}()
	return finishChan
}

func LoginAll() {
	var unverifiedPool = make(chan *Account, len(CookieManager.cookieAccounts))
	for _, account := range CookieManager.cookieAccounts {
		if account.id != "" {
			Console.Log(CATEGORY_ROBLOX, LOG_TRACE, "Account has ID")
		} else {
			unverifiedPool <- account
		}
	}

	for i := 1; i <= 10; i++ {
		go func() {
			var account *Account
			for {
				select {
				case account = <-unverifiedPool:
					for {
						success := <-Login(account)
						if success {
							break
						}
						time.Sleep(20 * time.Second)
					}
				default:
					return
				}
			}
		}()
	}
}