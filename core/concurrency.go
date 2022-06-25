package core

import (
	"fmt"
	. "kekops/ui"
	"sync"
)

const JobsPerProxy = 2

// Spreads cookies over the most optimal number of goroutines to execute tasks
func LaunchJob(fn func(*Proxy, chan string)) <-chan bool {
	finish := make(chan bool)
	go func() {
		var wg sync.WaitGroup
		cookieChan := make(chan string, len(CookieManager.cookies))
		for _, cookie := range CookieManager.cookies {
			cookieChan <- cookie
		}
		for j := 0; j < JobsPerProxy; j++ {
			for n := 0; n < len(Proxies); n++ {
				wg.Add(1)
				go func(sect int) {
					defer func() {
						if r := recover(); r != nil {
							Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Captured panic! %v", r))
						}
					}()
					fn(Proxies[sect], cookieChan)
					wg.Done()
				}(n)
			}
		}

		wg.Wait()
		finish <- true
	}()
	return finish
}

func LaunchJobLimited(fn func(*Proxy, chan string), count int) <-chan bool {
	finish := make(chan bool)
	go func() {
		var wg sync.WaitGroup
		cookieChan := make(chan string, len(CookieManager.cookies))
		for n := 0; n < count; n++ {
			cookieChan <- CookieManager.cookies[n]
		}
		for j := 0; j < JobsPerProxy; j++ {
			for n := 0; n < len(Proxies); n++ {
				wg.Add(1)
				go func(sect int) {
					defer func() {
						if r := recover(); r != nil {
							Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Captured panic! %v", r))
						}
					}()
					fn(Proxies[sect], cookieChan)
					wg.Done()
				}(n)
			}
		}

		wg.Wait()
		finish <- true
	}()
	return finish
}

func LaunchJobVerified(fn func(*Proxy, chan string)) <-chan bool {
	finish := make(chan bool)
	go func() {
		var wg sync.WaitGroup
		cookieChan := make(chan string, len(CookieManager.verifiedCookies))
		for _, cookie := range CookieManager.verifiedCookies {
			cookieChan <- cookie
		}
		for j := 0; j < JobsPerProxy; j++ {
			for n := 0; n < len(Proxies); n++ {
				wg.Add(1)
				go func(sect int) {
					defer func() {
						if r := recover(); r != nil {
							Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Captured panic! %v", r))
						}
					}()
					fn(Proxies[sect], cookieChan)
					wg.Done()
				}(n)
			}
		}

		wg.Wait()
		finish <- true
	}()
	return finish
}

func LaunchJobAccounts(fn func(*Proxy, chan *Account)) <-chan bool {
	finish := make(chan bool)
	go func() {
		var wg sync.WaitGroup
		cookieChan := make(chan *Account, len(CookieManager.cookieAccounts))
		for _, cookie := range CookieManager.cookieAccounts {
			cookieChan <- cookie
		}
		for j := 0; j < JobsPerProxy; j++ {
			for n := 0; n < len(Proxies); n++ {
				wg.Add(1)
				go func(sect int) {
					defer func() {
						if r := recover(); r != nil {
							Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Captured panic! %v", r))
						}
					}()
					fn(Proxies[sect], cookieChan)
					wg.Done()
				}(n)
			}
		}

		wg.Wait()
		finish <- true
	}()
	return finish
}
