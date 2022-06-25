package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	math2 "github.com/google/gxui/math"
	"io/ioutil"
	ui "kekops/termui/v3"
	. "kekops/ui"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Icon Data, Group Name, Group Description
const GroupData = "--UUDD-LRLR-BABA\r\n" +
"Content-Type: image/png\r\n" +
"Content-Disposition: form-data; name=\"icon\"; filename=\"icon.png\"\r\n" +
"\r\n" + "%s\r\n" +
"--UUDD-LRLR-BABA\r\n" +
"Content-Disposition: form-data; name=\"name\"\r\n" +
"\r\n" + "%s\r\n" +
"--UUDD-LRLR-BABA\r\n" +
"Content-Disposition: form-data; name=\"description\"\r\n" +
"\r\n" + "%s\r\n" +
"--UUDD-LRLR-BABA\r\n" +
"Content-Disposition: form-data; name=\"publicGroup\"\r\n" +
"\r\n" + "true\r\n" +
"--UUDD-LRLR-BABA--"

var TransTable = map[string]string{
	"a": "а",
	"A": "А",
	"b": "b",
	"B": "В",
	"c": "с",
	"C": "С",
	"d": "d",
	"D": "Ⅾ",
	"e": "е",
	"E": "Ε",
	"f": "f",
	"g": "ɢ",
	"G": "Ԍ",
	"H": "Н",
	"i": "і",
	"I": "І",
	"j": "ϳ",
	"J": "Ј",
	"k": "κ",
	"K": "K",
	"L": "Ꮮ",
	"m": "ⅿ",
	"M": "Μ",
	"N": "Ν",
	"o": "o",
	"O": "Ο",
	"p": "р",
	"P": "Р",
	"s": "ѕ",
	"S": "Ѕ",
	"T": "Τ",
	"u": "υ",
	"U": "Ս",
	"v": "ν",
	"V": "Ꮩ",
	"W": "Ꮃ",
	"x": "х",
	"X": "Х",
	"y": "у",
	"Y": "Υ",
	"Ζ": "z",
}

func SpoofName(name string) string {
	runes := []rune(name)
	pos := 0
	for {
		if pos > len(name) {
			break
		}
		char := string(runes[pos])
		if len(char) > 1 {
			pos++
			continue
		}
		if rep, ok := TransTable[char]; ok {
			runes[pos] = []rune(rep)[0]
			break
		}
		pos++
	}
	return string(runes)
}

type SpamPluginCloneTask struct {
	BaseTask

	Goal  uint64
	Count uint64

	TitleRepeats int
	DescRepeats  int
}

type GroupOutput struct {
	ID          int    `json:"id"`
}

func (s *SpamPluginCloneTask) Execute() {
	imp := s.TDestination.(*Import)
	data := imp.Data()
	if len(data) == 0 {
		return
	}

	s.Goal = 500
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


	info := <-FetchProductInfo(imp.Original)
	Console.Log(CATEGORY_MAIN, LOG_INFO, info.Creator.Name)
	var out = &GroupOutput{}
	var groupData []byte
	var spoofedName = info.Creator.Name
	TryGroup := func(cookie string) bool {
		xsrf := <-FetchXsrf(cookie)
		spoofedName = SpoofName(spoofedName)
		body := fmt.Sprintf(GroupData, imp.Icon(), spoofedName, fmt.Sprintf("%s Fan Club", info.Creator.Name))
		request, _ := http.NewRequest("POST", "https://groups.roblox.com/v1/groups/create", bytes.NewReader([]byte(body)))
		request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", Boundary))

		request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
		request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
		request.Header.Add("User-Agent", UserAgent)
		request.Header.Add("X-CSRF-TOKEN", xsrf)

		resp, err := Proxies[rand.Intn(len(Proxies))].client.Do(request)
		if err != nil {
			Console.Log(CATEGORY_ROBLOX, LOG_ERROR, fmt.Sprintf("Error sending group create request: %v", err))
			return false
		}
		groupData, _ = ioutil.ReadAll(resp.Body)
		Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Group data: %s", string(groupData)))
		if strings.Contains(string(groupData), "TooManyRequests") {
			time.Sleep(10 * time.Second)
			return false
		}
		if !strings.Contains(string(groupData), "taken") && !strings.Contains(string(groupData), "denied") {
			return true
		}
		return false
	}
	group := Tracker.FindGroup(info.Creator.Name)
	if group != nil {
		Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Using existing group %s", group.ID))
		out.ID, _ = strconv.Atoi(group.ID)
	} else {
		for {
			if TryGroup(BotHolder) {
				break
			}
		}
		error := json.Unmarshal(groupData, out)
		if error != nil {
			Console.Log(CATEGORY_ROBLOX, LOG_ERROR, fmt.Sprintf("Failed to unmarshal: %v", error))
			return
		}
		Tracker.AddGroup(info.Creator.Name, strconv.Itoa(out.ID))
	}

	<-LaunchJobLimited(func(proxy *Proxy, cookies chan string) {
		upload := proxy.TryRequest(GenerateBotUploadPluginSpam(BotHolder, imp.Name, data, 100, 1000, fmt.Sprintf("%d", out.ID)))
		if upload != "" {
			Console.Log(CATEGORY_ROBLOX, LOG_TRACE, upload)
		}
		if _, err := strconv.Atoi(upload); err != nil {
			return
		}
		for {
			resp := Proxies[rand.Intn(len(Proxies))].TryRequest(GenerateThumbnailRequest(BotHolder, upload, imp.Icon()))
			if strings.Contains(resp, "target") {
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Uploaded icon: %s", resp))
				break
			} else {
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Failed icon: %s", resp))
			}
		}
		info := <-FetchProductInfo(upload)
		if info == nil {
			return
		}
		
		atomic.AddUint64(&s.Count, 1)
	}, int(50))
}

func (b *SpamPluginCloneTask) Color() string {
	return "green"
}

func (b *SpamPluginCloneTask) Type() string {
	return "SPAM"
}
