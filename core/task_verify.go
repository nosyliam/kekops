package core

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"regexp"
	"time"

	//"regexp"

	//"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/thanhpk/randstr"
	"golang.org/x/net/websocket"
	//"io/ioutil"
	. "kekops/ui"
	//"net/http"
	//"strings"
	//"time"
)

type VerifyTask struct {
	BaseTask

	Goal  uint64
	Count uint64
}

func CreateInbox(address string) {
	url := fmt.Sprintf("https://mailsac.com/api/addresses/%s@topkek.gg", address)
	payload := strings.NewReader("{\"info\":\"string\",\"forward\":\"\",\"enablews\":true,\"webhook\":\"\",\"webhookSlack\":\"\",\"webhookSlackToFrom\":true}")


	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Mailsac-Key", "k_1f0ZrqNmOzEob3gAi3boTucBlJgaarPQfdjsw8CWdfM4lHg0e0f")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	Console.Log(CATEGORY_MAIN, LOG_INFO, string(body))
}

func DeleteInbox(address string) {
	url := fmt.Sprintf("https://mailsac.com/api/addresses/%s@topkek.gg", address)
	payload := strings.NewReader("{}")


	req, _ := http.NewRequest("DELETE", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("Mailsac-Key", "k_1f0ZrqNmOzEob3gAi3boTucBlJgaarPQfdjsw8CWdfM4lHg0e0f")

	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	Console.Log(CATEGORY_MAIN, LOG_INFO, string(body))
}

func readClientMessages(ws *websocket.Conn, incomingMessages chan string, finish chan bool) {
	for {
		select {
		case <-finish:
			return
		default:
			var message string
			// err := websocket.JSON.Receive(ws, &message)
			err := websocket.Message.Receive(ws, &message)
			if err != nil {
				Console.Log(CATEGORY_MAIN, LOG_TRACE, fmt.Sprintf("Websocket error: %s", err))
				return
			}
			incomingMessages <- message
		}
	}
}

func ListenForEmail(address string, finish chan bool) chan string {
	type Email struct {
		Text string `json:"text"`
	}
	ws, err := websocket.Dial(fmt.Sprintf("wss://sock.mailsac.com/incoming-messages?key=k_1f0ZrqNmOzEob3gAi3boTucBlJgaarPQfdjsw8CWdfM4lHg0e0f&addresses=%s@topkek.gg", address), "", "http://sock.mailsac.com/")
	if err != nil {
		Console.Log(CATEGORY_MAIN, LOG_ERROR, fmt.Sprintf("Dial failed: %s\n", err.Error()))
		return nil
	}
	incomingMessages := make(chan string)
	finishChan := make(chan bool)
	go readClientMessages(ws, incomingMessages, finishChan)
	mailChan := make(chan string)
	go func() {
		for {
			select {
			case <-finish:
				finishChan <- true
				return
			case message := <-incomingMessages:
				msg := &Email{}
				err := json.Unmarshal([]byte(message), msg)
				if err != nil {
					Console.Log(CATEGORY_ROBLOX, LOG_ERROR, fmt.Sprintf("Failed to receive message: %v", err))
				} else {
					mailChan <- msg.Text
				}
			}
		}
	}()
	return mailChan
}

func (v *VerifyTask) Execute() {
	Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Verifying %d cookies", len(CookieManager.cookieAccounts)))
	v.Goal = uint64(len(CookieManager.cookieAccounts))
	var unverifiedPool = make(chan *Account, len(CookieManager.cookieAccounts))
	accountMap := make(map[string]*Account)
	for _, account := range CookieManager.cookieAccounts {
		if CheckAccountVerified(account.cookie) {
			Console.Log(CATEGORY_ROBLOX, LOG_TRACE, "Account was found verified")
		} else {
			accountMap[account.user] = account
			unverifiedPool <- account
			break
		}
	}
	Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("%d accounts are unverified", len(unverifiedPool)))
	// for {
	var unverifiedCookies []*Account
	for n := 0; n < 5; n++ {
		var cookie *Account
		select {
		case cookie = <-unverifiedPool:
		default:
			continue
		}
		unverifiedCookies = append(unverifiedCookies, cookie)
	}
	verificationWorker := func(account *Account) {
	restart:
		Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Verifying account %s", account.user))
		inbox := randstr.Hex(10)
		CreateInbox(inbox)
		time.Sleep(1 * time.Second)
		finishChan := make(chan bool)
		incomingMessages := ListenForEmail(inbox, finishChan)
		<-incomingMessages
		proxy := Proxies[rand.Intn(len(Proxies))]
		resp := proxy.TryRequest(GenerateBotVerify(account.cookie, inbox+"@topkek.gg"))
		Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Verify response: %s", resp))
		if resp != "{}" {
			DeleteInbox(inbox)
			finishChan <- true
			time.Sleep(2 * time.Second)
			goto restart
		}
		for {
			select {
			case msg := <-incomingMessages:
				regex, _ := regexp.Compile("(https://www\\.roblox\\.com/account/settings/verify-email\\?ticket)=([^\\)]*)")
				matches := regex.FindStringSubmatch(msg)
				if len(matches) != 3 {
					Console.Log(CATEGORY_ROBLOX, LOG_ERROR, msg)
					DeleteInbox(inbox)
					finishChan <- true
					return
				}

				userReg, _ := regexp.Compile("(Roblox account <strong>)([^\\)\\n]*?)(</strong>)")
				userMatches := userReg.FindStringSubmatch(msg)
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, strings.Join(userMatches, ","))
				user := strings.TrimSpace(userMatches[2])
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("User: %s", user))

				req, _ := http.NewRequest("POST", "https://accountinformation.roblox.com/v1/email/verify",
					bytes.NewBuffer([]byte(fmt.Sprintf("{ticket: \"%s\"}", strings.TrimSpace(matches[2])))))
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				res, err := ProxylessClient.Do(req)
				xsrf := res.Header.Get("x-csrf-token")
				req, _ = http.NewRequest("POST", "https://accountinformation.roblox.com/v1/email/verify",
					bytes.NewBuffer([]byte(fmt.Sprintf("{ticket: \"%s\"}", strings.TrimSpace(matches[2])))))
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("XSRF: %s", xsrf))
				req.Header.Set("Content-Type", "application/json; charset=utf-8")
				req.Header.Add("X-CSRF-TOKEN", xsrf)
				res, err = ProxylessClient.Do(req)
				if err != nil {
					Console.Log(CATEGORY_ROBLOX, LOG_ERROR, fmt.Sprintf("Failed to verify: %v", err))
					finishChan <- true
					DeleteInbox(inbox)
					goto restart
				}
				defer res.Body.Close()
				body, _ := ioutil.ReadAll(res.Body)
				Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("Verification response: %s", body))

				if account, ok := accountMap[user]; ok {
					Login(account)
				} else {
					Console.Log(CATEGORY_ROBLOX, LOG_ERROR, fmt.Sprintf("Could not find user %s in account map.", user))
				}
			case <-time.After(5 * time.Second):
				finishChan <- true
				DeleteInbox(inbox)
				goto restart
			}
		}
	}

	verificationWorker(unverifiedCookies[0])

}

func (v *VerifyTask) Color() string {
	return "green"
}

func (v *VerifyTask) Type() string {
	return "VERIFY"
}