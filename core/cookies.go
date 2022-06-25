package core

import (
	"fmt"
	"io/ioutil"
	. "kekops/ui"
	"math/rand"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var CookieManager *KekCookieManager

type Account struct {
	user   string
	pass   string
	cookie string
	id     string
}

type KekCookieManager struct {
	cookies         []string
	verifiedCookies []string
	cookieAccounts  []*Account
}

func (k *KekCookieManager) Cookies(sect int, cpj int) []string {
	cookies := make([]string, cpj)
	for n := cpj; n > 0; n-- {
		cookie := k.Cookie((sect * cpj) + (n - 1))
		if cookie != "" {
			cookies[cpj-n] = cookie
		} else {
			break
		}
	}
	return cookies
}

func (k *KekCookieManager) Verified() string {
	// Return a random verified cookie
	rand.Seed(time.Now().Unix())
	if len(k.verifiedCookies) > 0 {
		return k.verifiedCookies[rand.Int31n(int32(len(k.verifiedCookies)-1))]
	} else {
		return ""
	}
}

func (k *KekCookieManager) Account() *Account {
	// Return a random account cookie
	rand.Seed(time.Now().Unix())
	for {
		if len(k.cookieAccounts) > 0 {
			acc := k.cookieAccounts[rand.Int31n(int32(len(k.cookieAccounts)-1))]
			if acc.id != "" {
				return acc
			}
		} else {
			return nil
		}
	}
}


func (k *KekCookieManager) Cookie(index int) string {
	if index >= len(k.cookies) {
		return ""
	}
	return k.cookies[index]
}

func (k *KekCookieManager) LoadCookies(file string, verify bool) {
	for {
		if Console != nil {
			break
		}
	}
	if len(Proxies) == 0 {
		Console.Log(CATEGORY_MAIN, LOG_ERROR, "No proxies were detected.\n KEKOPS cannot run without proxies.")
		return
	}
	if file, err := os.OpenFile("verified.txt", os.O_RDWR, 0755); err != nil {
		Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Unable to load verified cookies.\n Kekops will be unable to perform auto-toolbox.", err))
	} else {
		data, _ := ioutil.ReadAll(file)
		k.verifiedCookies = strings.SplitN(string(data), "\n", -1)
		for n, cookie := range k.verifiedCookies {
			k.verifiedCookies[n] = strings.TrimSpace(cookie)
		}
	}
	if file, err := os.OpenFile(file, os.O_RDWR, 0755); err != nil {
		Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Unable to load cookies: %v", err))
	} else {
		data, _ := ioutil.ReadAll(file)
		k.cookies = strings.SplitN(string(data), "\n", -1)
		k.cookieAccounts = make([]*Account, len(k.cookies))
		for n, cookie := range k.cookies {
			k.cookies[n] = cookie
			cookie := strings.SplitN(cookie, ":", -1)
			if len(cookie) > 3 {
				k.cookies[n] = strings.TrimSpace(strings.Join(cookie[2:4], ":"))
				k.cookieAccounts[n] = &Account{user: cookie[0], pass: cookie[1], cookie:  strings.TrimSpace(strings.Join(cookie[2:4], ":"))}
				if len(cookie) == 6 {
					k.cookieAccounts[n].id = strings.TrimSpace(strings.Join(cookie[4:6], ":"))
				}
			}
		}
		if !verify {
			return
		}
		// Launch a task which verifies all cookies
		var checkedCookies uint64
		log := Console.Log(CATEGORY_MAIN, LOG_INFO, fmt.Sprintf("Verified 0/%d cookies...", len(k.cookies)))
		var workingCookies []string
		testInfo := <-FetchProductInfo("6267075455")
		if testInfo == nil {
			return
		}
		<-LaunchJob(func(proxy *Proxy, cookies chan string) {
			for {
				if len(cookies) == 0 {
					return
				}
				cookie := <-cookies
				sale := proxy.TryRequest(GenerateBotTakeRequestPlugin(cookie, *testInfo))
				if sale != "" {
					if !strings.Contains(sale, "denied") {
						workingCookies = append(workingCookies, cookie)
					}
				}
				atomic.AddUint64(&checkedCookies, 1)
				log.Update(fmt.Sprintf("Verified %d/%d cookies...", atomic.LoadUint64(&checkedCookies), len(k.cookies)))
			}
		})
		Console.Log(CATEGORY_MAIN, LOG_INFO, fmt.Sprintf("Detected [%d](fg:yellow) valid cookies.", len(workingCookies)))
		k.cookies = workingCookies
	}
}

func (k *KekCookieManager) Save(out string) {
	ioutil.WriteFile(out, []byte(strings.Join(k.cookies, "\n")), os.FileMode(077))
}

func (k *KekCookieManager) SaveAccounts(out string) {
	accounts := make([]string, len(k.cookieAccounts))
	for n, cookie := range k.cookieAccounts {
		accounts[n] = fmt.Sprintf("%s:%s:%s:%s", cookie.user, cookie.pass, cookie.cookie, cookie.id)
	}
	ioutil.WriteFile(out, []byte(strings.Join(accounts, "\n")), os.FileMode(077))
}


func init() {
	CookieManager = &KekCookieManager{}
}
