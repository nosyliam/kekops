package core

import (
	"fmt"
	. "kekops/ui"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

var TitleRepeats, DescRepeats = 1, 10

func RegisterCommands() {
	defer func() {
		if r := recover(); r != nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Caught panic!")
		}
	}()

	for {
		if Console != nil {
			break
		}
	}
	Console.RegisterCommand("help", func(args ...string) {
		Console.Log(CATEGORY_CONSOLE, LOG_NOLEVEL, `[KEKOPS COMMAND LIST](fg:magenta)
 - newtask [type] [args...]
      bot       [id] [count]
      like      [id]
      favorite  [id]
      spam      [import]
      scrape    [query] [count]
 - canceltask
 - spamsettings [titleRepeats] [descRepeats]
 - importmodel  [term] (offset)
 - importplugin [id]
 - importlist
 - name [import] [name]
 - removeimport [import]
 - frontpage    [term] (offset)
 - toolbox      [import]
 - loadproxies  [file]
 - loadcookies  [file]
 - loadbackdoor [file]
 - toggletrace
`)
	})
	Console.RegisterCommand("spamsettings", func(args ...string) {
		if len(args) != 2 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
		}
		title, err := strconv.Atoi(args[0])
		if err != nil {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #1:\n Must be an integer")
			return
		}
		desc, err := strconv.Atoi(args[1])
		if err != nil {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #2:\n Must be an integer")
			return
		}
		TitleRepeats, DescRepeats = title, desc
	})
	Console.RegisterCommand("newtask", func(args ...string) {
		defer func() {
			if r := recover(); r != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
			}
		}()
		var useHolder bool
		switch strings.ToLower(args[0]) {
		case "verify":
			Scheduler.Queue <- &VerifyTask{}
		case "bot":
			var count int
			id, err := strconv.Atoi(args[1])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #3:\n Destination must be an integer")
				return
			}
			count, err = strconv.Atoi(args[2])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #4:\n Count must be an integer")
				return
			}
			Scheduler.Queue <- &BotTask{BaseTask: BaseTask{TDestination: id}, Goal: uint64(count)}
		case "insert":
			var count int
			id, err := strconv.Atoi(args[1])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #3:\n Destination must be an integer")
				return
			}
			count, err = strconv.Atoi(args[2])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #4:\n Count must be an integer")
				return
			}
			Scheduler.Queue <- &InsertTask{BaseTask: BaseTask{TDestination: id}, Goal: uint64(count)}
		case "scrape":
			var count int
			var err error
			query := args[1:]
			if count, err = strconv.Atoi(args[len(args)-1]); err == nil {
				query = args[1 : len(args)-1]
			}
			if len(query) == 0 {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
				return
			}
			queryString := strings.Join(query, " ")
			Scheduler.Queue <- &ScrapeTask{BaseTask: BaseTask{TDestination: queryString}, Offset: uint64(count), Goal: uint64(count)}
		case "scrapeuser":
			var count int
			var err error
			query := args[1:]
			if count, err = strconv.Atoi(args[len(args)-1]); err == nil {
				query = args[1 : len(args)-1]
			}
			if len(query) == 0 {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
				return
			}
			queryString := strings.Join(query, " ")
			Scheduler.Queue <- &ScrapeTask{BaseTask: BaseTask{TDestination: queryString}, Offset: uint64(count), Goal: uint64(count), Creator: true}
		case "analytics":
			var count int
			id, err := strconv.Atoi(args[1])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #3:\n Destination must be an integer")
				return
			}
			count, err = strconv.Atoi(args[2])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #4:\n Count must be an integer")
				return
			}
			Scheduler.Queue <- &AnalyticsTask{BaseTask: BaseTask{TDestination: id}, Goal: uint64(count), QueryName: strings.Join(args[3:], " ")}
		case "spam":
			imp := Tracker.Find(strings.Join(args[1:], " "))
			if imp == nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Import could not be found.")
				return
			}
			switch imp.Type {
			case IMPORT_MODEL:
				Scheduler.Queue <- &SpamTask{BaseTask: BaseTask{TDestination: imp, TDisplayDestination: imp.Name}, TitleRepeats: TitleRepeats, DescRepeats: DescRepeats}
			case IMPORT_PLUGIN:
				Scheduler.Queue <- &SpamPluginTask{BaseTask: BaseTask{TDestination: imp, TDisplayDestination: imp.Name}, TitleRepeats: TitleRepeats, DescRepeats: DescRepeats}
			}
		case "upload":
			imp := Tracker.Find(strings.Join(args[1:], " "))
			if imp == nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Import could not be found.")
				return
			}
			switch imp.Type {
			case IMPORT_MODEL:
				Scheduler.Queue <- &UploadTask{BaseTask: BaseTask{TDestination: imp}}
			case IMPORT_PLUGIN:
			}
		case "clone":
			imp := Tracker.Find(strings.Join(args[1:], " "))
			if imp == nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Import could not be found.")
				return
			}
			Scheduler.Queue <- &SpamPluginCloneTask{BaseTask: BaseTask{TDestination: imp, TDisplayDestination: imp.Name}, TitleRepeats: TitleRepeats, DescRepeats: DescRepeats}
		case "like":
			id, err := strconv.Atoi(args[1])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #3:\n Destination must be an integer")
				return
			}
			Scheduler.Queue <- &LikeTask{BaseTask: BaseTask{TDestination: id}}
		case "dislike":
			id, err := strconv.Atoi(args[1])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #3:\n Destination must be an integer")
				return
			}
			Scheduler.Queue <- &DislikeTask{BaseTask: BaseTask{TDestination: id}}
		case "favorite":
			id, err := strconv.Atoi(args[1])
			if err != nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #3:\n Destination must be an integer")
				return
			}
			Scheduler.Queue <- &FavoriteTask{BaseTask: BaseTask{TDestination: id}}
		case "autotb":
			imp := Tracker.Find(strings.Join(args[1:], " "))
			if imp == nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Import could not be found.")
				return
			}
			Scheduler.Queue <- &AutoToolboxTask{BaseTask: BaseTask{TDestination: imp, TDisplayDestination: imp.Name}}
		case "autofpgroup":
			useHolder = true
			fallthrough
		case "autofp":
			imp := Tracker.Find(strings.Join(args[1:], " "))
			if imp == nil {
				Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Import could not be found.")
				return
			}
			switch imp.Type {
			case IMPORT_MODEL:
				Scheduler.Queue <- &AutoFrontPageTask{BaseTask: BaseTask{TDestination: imp, TDisplayDestination: imp.Name}, UseHolder: useHolder}
			case IMPORT_PLUGIN:
				Scheduler.Queue <- &AutoPluginFrontPageTask{BaseTask: BaseTask{TDestination: imp, TDisplayDestination: imp.Name}, UseHolder: useHolder}

			}
		default:
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Unknown task type.")
		}
	})
	Console.RegisterCommand("botall", func(args ...string) {
		BotAll()
	})
	Console.RegisterCommand("clearmodels", func(args ...string) {
		Tracker.Models = []int{}
		Tracker.Save()
	})
	Console.RegisterCommand("importmodel", func(args ...string) {
		var offset int
		var err error
		query := args
		if offset, err = strconv.Atoi(args[len(args)-1]); err == nil {
			query = args[0 : len(args)-1]
		}
		if len(query) == 0 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
			return
		}
		queryString := strings.Join(query, " ")
		Operations.ImportModel(queryString, offset)
	})
	Console.RegisterCommand("importplugin", func(args ...string) {
		if _, err := strconv.Atoi(args[0]); err == nil {
			if args[len(args)-1] == "noverlay" {
				Operations.ImportPlugin(args[0], false, strings.Join(args[1:len(args)-1], " "))
			} else {
				Operations.ImportPlugin(args[0], true, strings.Join(args[1:], " "))
			}
		} else {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid argument #1: ID should be an integer")
			return
		}
	})
	Console.RegisterCommand("removeimport", func(args ...string) {
		Tracker.Remove(strings.Join(args, " "))
	})
	Console.RegisterCommand("name", func(args ...string) {
		if len(args) < 2 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
			return
		}
		Tracker.UpdateName(args[0], strings.Join(args[1:], " "))
	})
	Console.RegisterCommand("importlist", func(args ...string) {
		Tracker.List()
	})
	Console.RegisterCommand("savecookies", func(args ...string) {
		if len(args) != 1 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
			return
		}
		CookieManager.Save(args[0])
	})
	Console.RegisterCommand("saveaccounts", func(args ...string) {
		if len(args) != 1 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
			return
		}
		CookieManager.SaveAccounts(args[0])
	})
	Console.RegisterCommand("frontpage", func(args ...string) {
		Tracker.UpdateMode(strings.Join(args, " "), IMPORT_LEVEL_FRONTPAGE)
	})
	Console.RegisterCommand("toolbox", func(args ...string) {
		Tracker.UpdateMode(strings.Join(args, " "), IMPORT_LEVEL_TOOLBOX)
	})
	Console.RegisterCommand("clear", func(args ...string) {
		Console.Logs = []*Log{}
		go Console.Render()
	})
	Console.RegisterCommand("quit", func(args ...string) {
		go func() {
			for n := 0; n < 50; n++ {
				ExitChan <- true
			}
		}()
	})
	Console.RegisterCommand("loadcookies", func(args ...string) {
		if len(args) == 0 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
			return
		}
		CookieManager.LoadCookies(args[0], true)
	})
	Console.RegisterCommand("loadproxies", func(args ...string) {
		if len(args) == 0 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
			return
		}
		LoadProxies(args[0])
	})
	Console.RegisterCommand("loadbackdoor", func(args ...string) {
		if len(args) != 1 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
		}
		Operations.LoadBackdoor(args[0])
	})
	Console.RegisterCommand("loadbackdoorplugin", func(args ...string) {
		if len(args) != 1 {
			Console.Log(CATEGORY_CONSOLE, LOG_ERROR, "Invalid number of arguments.")
		}
		Operations.LoadBackdoorPlugin(args[0])
	})
	Console.RegisterCommand("canceltask", func(args ...string) {
		if TaskList.CurrentTask != nil {
			TaskList.CurrentTask.Quit()
		}
	})
	Console.RegisterCommand("toggletrace", func(args ...string) {
		TraceEnabled = !TraceEnabled
	})
	Console.RegisterCommand("dumpheap", func(args ...string) {
		go func() {
			for {
				file, _ := os.Create("heap.txt")
				pprof.WriteHeapProfile(file)
				file.Sync()
				file.Close()
				time.Sleep(10 * time.Second)
			}
		}()
	})
	Console.RegisterCommand("cpuprofile", func(args ...string) {
		f, _ := os.Create("cpu.txt")
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	})
	Console.RegisterCommand("removefp", func(args ...string) {
		<-LaunchJobVerified(func(proxy *Proxy, cookies chan string) {
			var cookie string
			select {
			case cookie = <-cookies:
			default:
				return
			}
			proxy.TryRequest(GenerateBotDisable(cookie, args[0]))
		})
	})
	Console.RegisterCommand("testencode", func(args ...string) {
		Console.Log(CATEGORY_MAIN, LOG_INFO, fmt.Sprintf("Output: %s", Operations.EncodeRequire(args[0])))
	})
	Console.RegisterCommand("logintest", func(args ...string) {
		Login(CookieManager.cookieAccounts[0])
	})
	Console.RegisterCommand("login", func(args ...string) {
		LoginAll()
	})
	Console.RegisterCommand("list", func(args ...string) {
		Tracker.ListRequires()
	})
	Console.RegisterCommand("redirect", func(args ...string) { // redirect 6803973136 5700000000
		newReq := Operations.GenerateRequireModel()
		min, err := strconv.Atoi(args[len(args)-1])
		for _, req := range Tracker.Requires {
			if err == nil {
				id, _ := strconv.Atoi(req.ID)
				//Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("%d, %d", id, min))
				if id < min {
					continue
				}
			}
			Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Redirecting module %s: %s", req.ID,
				Proxies[0].TryRequest(GenerateBotUploadRequireReupload(req.Cookie, req.ID, newReq))))
		}
	})
}
