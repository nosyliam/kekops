package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/gxui/math"
	"io/ioutil"
	. "kekops/ui"
	math2 "math"
	"math/rand"
	"net/http"
	gourl "net/url"
	"os"
	"strings"
	"time"
)

const (
	UserAgent = "Roblox"
)

const likeRatio = 10

const Boundary = `UUDD-LRLR-BABA`
const FormData = "--%s\r\n" +
	"Content-Type: image/%s\r\n" +
	"Content-Disposition: form-data; filename=\"%s\"; name=\"request.files\"\r\n" +
	"\r\n" +
	"%s\r\n" +
	"--%s--\r\n"

func shouldLike() string {
	rand.Seed(time.Now().Unix())
	if rand.Intn(likeRatio) == 1 {
		return "false"
	}
	return "true"
}

var ProxylessClient *http.Client

type ProductInfo struct {
	TargetID    int    `json:"TargetId"`
	ProductType string `json:"ProductType"`
	AssetID     int    `json:"AssetId"`
	ProductID   int    `json:"ProductId"`
	Name        string `json:"Name"`
	Description string `json:"Description"`
	AssetTypeID int    `json:"AssetTypeId"`
	Creator     struct {
		ID   int    `json:"Id"`
		Name string `json:"Name"`
	} `json:"Creator"`
	IconImageAssetID       int         `json:"IconImageAssetId"`
	Created                time.Time   `json:"Created"`
	Updated                time.Time   `json:"Updated"`
	PriceInRobux           int         `json:"PriceInRobux"`
	PriceInTickets         interface{} `json:"PriceInTickets"`
	Sales                  int         `json:"Sales"`
	IsNew                  bool        `json:"IsNew"`
	IsForSale              bool        `json:"IsForSale"`
	IsPublicDomain         bool        `json:"IsPublicDomain"`
	IsLimited              bool        `json:"IsLimited"`
	IsLimitedUnique        bool        `json:"IsLimitedUnique"`
	Remaining              interface{} `json:"Remaining"`
	MinimumMembershipLevel int         `json:"MinimumMembershipLevel"`
}

type ToolboxSearch struct {
	TotalResults       int         `json:"totalResults"`
	FilteredKeyword    string      `json:"filteredKeyword"`
	PreviousPageCursor interface{} `json:"previousPageCursor"`
	NextPageCursor     string      `json:"nextPageCursor"`
	Data               []struct {
		ID       int    `json:"id"`
		ItemType string `json:"itemType"`
	} `json:"data"`
}

type ToolboxSuggestions struct {
	Data []struct {
		Query string      `json:"Query"`
	} `json:"Data"`
}

type ThumbnailData struct {
	Data []struct {
		TargetID int    `json:"targetId"`
		State    string `json:"state"`
		ImageURL string `json:"imageUrl"`
	} `json:"data"`
}

func SearchToolbox(query string, count int) <-chan *ToolboxSearch {
	out := make(chan *ToolboxSearch)
	go func() {
		var search *ToolboxSearch = nil
		try := func() bool {
			var cursor = ""
			for i := 0; i < int(math2.Round(math2.Ceil(float64(count) / 30))); i++ {
				account := CookieManager.Account()
				req, err := http.NewRequest("GET", fmt.Sprintf("https://apis.roblox.com/toolbox-service/v1/Models?keyword=%s&creatorTargetId=1580935894&cursor=%s",
					gourl.QueryEscape(query), cursor), nil)
				req.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: account.cookie})
				req.AddCookie(&http.Cookie{Name: ".RBXID", Value: account.id})
				resp, err := ProxylessClient.Do(req)
				if err != nil {
					return false
				}
				defer resp.Body.Close()

				data, _ := ioutil.ReadAll(resp.Body)
				if search == nil {
					search = &ToolboxSearch{}
					err = json.Unmarshal(data, search)
					cursor = search.NextPageCursor
				} else {
					var temp = &ToolboxSearch{}
					_ = json.Unmarshal(data, temp)
					search.Data = append(search.Data, temp.Data...)
					cursor = temp.NextPageCursor
				}
				if err != nil {
					Console.Log(CATEGORY_OPERATIONS, LOG_WARN, fmt.Sprintf("Failed to execute toolbox search: %v", err))
					return false
				}
			}
			return true
		}
		for n := 0; n < 2; n++ {
			if try() {
				out <- search
				break
			}
		}
		out <- nil
	}()
	return out
}

func SearchToolboxCreator(id string, count int) <-chan *ToolboxSearch {
	out := make(chan *ToolboxSearch)
	go func() {
		var search *ToolboxSearch = nil
		try := func() bool {
			var cursor = ""
			Console.Log(CATEGORY_ROBLOX, LOG_INFO, fmt.Sprintf("%d", int(math2.Round(math2.Ceil(float64(count) / 30)))))
			for i := 0; i < int(math2.Round(math2.Ceil(float64(count) / 30))); i++ {
				account := CookieManager.Account()
				req, err := http.NewRequest("GET", fmt.Sprintf("https://apis.roblox.com/toolbox-service/v1/Models?creatorTargetId=%s&cursor=%s",
					id, cursor), nil)
				req.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: account.cookie})
				req.AddCookie(&http.Cookie{Name: ".RBXID", Value: account.id})
				resp, err := ProxylessClient.Do(req)
				if err != nil {
					return false
				}
				defer resp.Body.Close()

				data, _ := ioutil.ReadAll(resp.Body)
				if search == nil {
					search = &ToolboxSearch{}
					err = json.Unmarshal(data, search)
					cursor = search.NextPageCursor
				} else {
					var temp = &ToolboxSearch{}
					_ = json.Unmarshal(data, temp)
					search.Data = append(search.Data, temp.Data...)
					cursor = temp.NextPageCursor
				}
				if err != nil {
					Console.Log(CATEGORY_OPERATIONS, LOG_WARN, fmt.Sprintf("Failed to execute toolbox search: %v", err))
					return false
				}
			}
			return true
		}
		for n := 0; n < 2; n++ {
			if try() {
				out <- search
				break
			}
		}
		out <- nil
	}()
	return out
}

func FetchSuggestions(query string) <-chan *ToolboxSuggestions {
	out := make(chan *ToolboxSuggestions)
	go func() {
		var search *ToolboxSuggestions = &ToolboxSuggestions{}
		try := func() bool {
			account := CookieManager.Account()
			req, err := http.NewRequest("GET", fmt.Sprintf("https://apis.roblox.com/autocomplete-studio/v2/suggest?prefix=%s&limit=100&cat=model",
				gourl.QueryEscape(query)), nil)
			req.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: account.cookie})
			req.AddCookie(&http.Cookie{Name: ".RBXID", Value: account.id})
			resp, err := ProxylessClient.Do(req)
			if err != nil {
				return false
			}
			defer resp.Body.Close()

			data, _ := ioutil.ReadAll(resp.Body)
			error := json.Unmarshal(data, search)
			if error != nil {
				Console.Log(CATEGORY_OPERATIONS, LOG_WARN, fmt.Sprintf("Failed to execute toolbox search suggestions: %v", error))
				return false
			}
			return true
		}
		for n := 0; n < 2; n++ {
			if try() {
				out <- search
				break
			}
		}
		out <- search
	}()
	return out
}

func FetchAssetThumbnail(id string) <-chan []byte {
	out := make(chan []byte)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Caught panic!")
			}
		}()

		var data []byte
		var Thumbnail = &ThumbnailData{}
		var try func() bool
		try = func() bool {
			time.Sleep(300 * time.Millisecond)
			req, err := http.NewRequest("GET", fmt.Sprintf("https://thumbnails.roblox.com/v1/assets?assetIds=%s&size=420x420&format=Png&isCircular=false", id), nil)
			if err != nil {
				return false
			}
			resp, err := ProxylessClient.Do(req)
			if err != nil {
				return false
			}
			time.Sleep(1000 * time.Millisecond)
			if resp.Body != nil {
				defer resp.Body.Close()
			}

			respData, _ := ioutil.ReadAll(resp.Body)
			err = json.Unmarshal(respData, Thumbnail)
			if err != nil {
				return false
			}
			if Thumbnail.Data[0].State == "Completed" {
				Console.Log(CATEGORY_OPERATIONS, LOG_TRACE, fmt.Sprintf("URL: %s", Thumbnail.Data[0].ImageURL))
				req, err := http.NewRequest("GET", Thumbnail.Data[0].ImageURL, nil)
				if err != nil {
					return false
				}
				resp, err = ProxylessClient.Do(req)
				if err != nil {
					return false
				}
				defer resp.Body.Close()
				data, _ = ioutil.ReadAll(resp.Body)
				return true
			} else {
				return try()
			}
			return true
		}
		for n := 0; n < 2; n++ {
			if try() {
				out <- data
				break
			}
		}
		out <- []byte("")
	}()
	return out
}

func FetchAsset(id string) <-chan []byte {
	out := make(chan []byte)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Caught panic!")
			}
		}()

		var data []byte
		try := func() bool {
			req, err := http.NewRequest("GET", fmt.Sprintf("https://assetdelivery.roblox.com/v1/asset?id=%s", id), nil)
			if err != nil {
				return false
			}
			resp, err := ProxylessClient.Do(req)
			if err != nil {
				return false
			}
			if resp.Body != nil {
				defer resp.Body.Close()
			}

			data, _ = ioutil.ReadAll(resp.Body)
			return true
		}
		for n := 0; n < 2; n++ {
			if try() {
				out <- data
				break
			}
		}
		out <- []byte("")
	}()
	return out
}

func FetchProductInfo(id string) <-chan *ProductInfo {
	out := make(chan *ProductInfo)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Console.Log(CATEGORY_MAIN, LOG_ERROR, "Captured panic!")
			}
		}()
		var product = &ProductInfo{}
		try := func() bool {
			req, err := http.NewRequest("GET", fmt.Sprintf("https://api.roblox.com/marketplace/productinfo?assetId=%s", id), nil)
			if err != nil {
				Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("ProductInfo fetch error: %v", err))
				return false
			}
			resp, err := Proxies[rand.Intn(len(Proxies)-1)].client.Do(req)
			if err != nil {
				return false
			}
			if resp.Body != nil {
				defer resp.Body.Close()
			}

			data, _ := ioutil.ReadAll(resp.Body)
			error := json.Unmarshal(data, product)
			if error != nil {
				return false
			}
			return true
		}
		for {
			if try() {
				out <- product
				break
			}
		}
		out <- nil
	}()
	return out
}

func FetchXsrf(cookie string) <-chan string {
	out := make(chan string)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Console.Log(CATEGORY_MAIN, LOG_ERROR, "Captured panic!")
			}
		}()
		var xsrf = ""
		try := func() bool {
			req, err := http.NewRequest("POST", "https://auth.roblox.com/v2/logout", nil)
			if err != nil {
				ErrorChart.IncrementHTTP()
				return false
			}
			req.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
			req.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
			resp, err := ProxylessClient.Do(req)
			if err != nil {
				ErrorChart.IncrementHTTP()
				return false
			}
			if resp.Body != nil {
				defer resp.Body.Close()
			}

			xsrf = resp.Header.Get("x-csrf-token")
			return true
		}
		for {
			if try() {
				out <- xsrf
				break
			} else {
			//	Console.Log(CATEGORY_MAIN, LOG_TRACE, "XSRF Failed")
			}
		}
	}()
	return out
}

func FetchUploadXSRF(cookie string) <-chan string {
	out := make(chan string)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				Console.Log(CATEGORY_MAIN, LOG_ERROR, "Captured panic!")
			}
		}()
		var xsrf = ""
		try := func() bool {
			req, err := http.NewRequest("POST", "https://data.roblox.com/Data/Upload.ashx", nil)
			if err != nil {
				ErrorChart.IncrementHTTP()
				return false
			}

			req.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
			req.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
			req.Header.Add("User-Agent", UserAgent)

			resp, _ := ProxylessClient.Do(req)
			if resp.Body != nil {
				defer resp.Body.Close()
			}
			xsrf = resp.Header.Get("x-csrf-token")
			return true
		}
		for {
			if try() {
				out <- xsrf
				break
			} else {
				//	Console.Log(CATEGORY_MAIN, LOG_TRACE, "XSRF Failed")
			}
		}
	}()
	return out
}


func GenerateBotVerify(cookie string, email string) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	request, _ := http.NewRequest("POST", "https://accountsettings.roblox.com/v1/email",
		bytes.NewReader([]byte(fmt.Sprintf("emailAddress=%s&password=", gourl.QueryEscape(email)))))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func CheckAccountVerified(cookie string) bool {
	type Response struct {
		EmailAddress string `json:"emailAddress"`
		Verified     bool   `json:"verified"`
	}
	var response *Response = &Response{}
	request, _ := http.NewRequest("GET", "https://accountsettings.roblox.com/v1/email", bytes.NewReader([]byte("")))
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	resp, err := ProxylessClient.Do(request)
	if err != nil {
		Console.Log(CATEGORY_ROBLOX, LOG_TRACE, err.Error())
		ErrorChart.IncrementHTTP()
		return false
	}
	data, _ := ioutil.ReadAll(resp.Body)
	Console.Log(CATEGORY_ROBLOX, LOG_TRACE, string(data))
	err = json.Unmarshal(data, response)
	if err == nil {
		return response.Verified
	} else {
		return CheckAccountVerified(cookie)
	}
	return false
}


func GenerateBotTakeRequest(cookie string, info ProductInfo) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://economy.roblox.com/v2/user-products/%d/purchase", info.ProductID),
		bytes.NewReader([]byte("")))


	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateBotTakeRequestPlugin(cookie string, info ProductInfo) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	body := []byte(fmt.Sprintf("{\"expectedCurrency\":\"1\",\"expectedPrice\":0,\"expectedSellerId\":\"%s\"}", info.Creator.ID))
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://economy.roblox.com/v1/purchases/products/%d", info.ProductID),
		bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateBotRemoveRequest(cookie string, id string) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://www.roblox.com/asset/delete-from-inventory?assetId=%s", id), bytes.NewReader([]byte("")))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateBotUpload(cookie string, name string, data []byte, useRepeats bool, useHolder bool) *http.Request {
	xsrf := <-FetchUploadXSRF(cookie)
	outName := name
	if useRepeats {
		for {
			if len(outName+" "+name) <= 51 {
				outName += name
			} else {
				break
			}
		}
	}

	var desc string
	if !useRepeats {
		desc = ""
	} else {
		desc = strings.Repeat(name, 10)
		desc = string([]byte(desc[0:math.Min(len(desc), 200)]))
	}
	url := "https://data.roblox.com/Data/Upload.ashx?"
	url += "assetId=0&"
	url += "type=Model&"
	url += "name=" + gourl.QueryEscape(outName) + "&"
	url += "description=" + gourl.QueryEscape(desc) + "&"
	url += "genreTypeId=0&"
	url += "ispublic=True&"
	url += "allowComments=True&"
	if useHolder {
		url += "groupId=9019868"
	} else {
		url += "groupId="
	}

	request, _ := http.NewRequest("POST", url, bytes.NewReader(data))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: BotHolder})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateBotUploadPluginSpam(cookie string, name string, data []byte, titleRepeat, descRepeat int, group string) *http.Request {
	xsrf := <-FetchUploadXSRF(cookie)
	//name += " "
	outName := name + " "
	//outName += name
	/*if titleRepeat != 1 {
		for {
			if len(outName+" "+name) <= 51 {
				outName += name
			} else {
				break
			}
		}
	}*/
	name += " "
	var desc string
	if descRepeat == 0 {
		desc = ""
	} else {
		/*for {
			if len(desc+" "+name) <= 700 {
				desc += name
			} else {
				break
			}
		}*/
		desc = strings.Repeat(name, descRepeat)
		desc = "Plugin Archived 2021\nArchived plugins are virus-free and occupy the library to prevent the spread of viruses. Thank you for using archived plugins.\n\n" + string([]byte(desc[0:math.Min(len(desc), 700)]))
		desc = string([]byte(desc[0:math.Min(len(desc), 700)]))
	}
	url := "https://data.roblox.com/Data/Upload.ashx?"
	url += "assetId=0&"
	url += "type=Plugin&"
	url += "name=" + gourl.QueryEscape(outName) + "&"
	url += "description=" + gourl.QueryEscape(desc) + "&"
	url += "genreTypeId=0&"
	url += "ispublic=True&"
	url += "allowComments=True&"
	url += "groupId=" + group

	request, _ := http.NewRequest("POST", url, bytes.NewReader(data))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	request.Header.Add("User-Agent", UserAgent)
	return request
}

func GenerateBotUploadPlugin(cookie string, data []byte, useHolder bool) *http.Request {
	url := "https://data.roblox.com/Data/Upload.ashx?"
	url += "assetId=0&"
	url += "type=Plugin&"
	url += "name=Plugin&"
	url += "description=Plugin&"
	url += "genreTypeId=0&"
	url += "ispublic=True&"
	url += "allowComments=False&"
	if useHolder {
		url += "groupId=9019868"
	} else {
		url += "groupId="
	}

	request, _ := http.NewRequest("POST", url, bytes.NewReader(data))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	return request
}

func GenerateBotUploadRequire(cookie string, id int) *http.Request {
	// Upload require with lua source
	xsrf := <-FetchUploadXSRF(cookie)
	url := "https://data.roblox.com/Data/Upload.ashx?"
	url += "assetId=0&"
	url += "type=Lua&"
	url += "name=MeshLoader&"
	url += "description=Module&"
	url += "genreTypeId=0&"
	url += "ispublic=True&"
	url += "allowComments=True&"
	url += "groupId="

	req := Operations.GenerateRequire(int64(id))
	request, _ := http.NewRequest("POST", url, bytes.NewReader(req))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateBotUploadRequireModel(cookie string) *http.Request {
	// Upload require with lua source
	xsrf := <-FetchUploadXSRF(cookie)
	url := "https://data.roblox.com/Data/Upload.ashx?"
	url += "assetId=0&"
	url += "type=Model&"
	url += "name=MeshLoader&"
	url += "description=Module&"
	url += "genreTypeId=0&"
	url += "ispublic=True&"
	url += "allowComments=True&"
	url += "groupId="

	req := Operations.GenerateRequireModel()
	request, _ := http.NewRequest("POST", url, bytes.NewReader(req))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateHash(asset string, cookie string) string {
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://assetdelivery.roblox.com/v1/assetId/%s", asset), nil)
	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	req, _ := ProxylessClient.Do(request)
	defer req.Body.Close()
	loc, _ := ioutil.ReadAll(req.Body)
	type resp struct {
		Location string `json:"location"`
	}


	ioutil.WriteFile("loc.txt", []byte(loc), os.FileMode(077))

	var data resp
	err := json.Unmarshal([]byte(loc), &data)
	if err != nil {
		return ""
	}

	dt := strings.SplitN(data.Location, "/", -1)
	if strings.Contains(data.Location, "contentstore") {
		dt = strings.SplitN(data.Location, "=", -1)
	}
	return dt[len(dt)-1]
}



/*
func GenerateBotUploadRequire(cookie string) *http.Request {
	xsrf := <-FetchUploadXSRF(cookie)
	url := "https://data.roblox.com/Data/Upload.ashx?"
	url += "assetId=0&"
	url += "type=Model&"
	url += "name=MeshLoader&"
	url += "description=Module&"
	url += "genreTypeId=0&"
	url += "ispublic=True&"
	url += "allowComments=True&"
	url += "groupId="

	req := Operations.GenerateRequire()
	request, _ := http.NewRequest("POST", url, bytes.NewReader(req))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}*/

func GenerateBotUploadRequireReupload(cookie string, id string, data []byte) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	url := "https://data.roblox.com/Data/Upload.ashx?"
	url += "assetId=" + gourl.QueryEscape(id) + "&"
	url += "type=Model&"
	url += "name=Module&"
	url += "description=Module&"
	url += "genreTypeId=0&"
	url += "ispublic=True&"
	url += "allowComments=False&"
	url += "groupId="

	request, _ := http.NewRequest("POST", url, bytes.NewReader(data))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	request.Header.Add("User-Agent", UserAgent)
	return request
}


func GenerateBotInsert(cookie *Account, id string) *http.Request {
	xsrf := <-FetchXsrf(cookie.cookie)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://apis.roblox.com/toolbox-service/v1/insert/asset/%s", id), bytes.NewReader([]byte("")))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie.cookie})
	request.AddCookie(&http.Cookie{Name: ".RBXID", Value: cookie.id})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateBotLike(cookie string, id string) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://www.roblox.com/voting/vote?assetId=%s&vote=true", id), bytes.NewReader([]byte("")))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: fmt.Sprintf("browserid=%d", rand.Int31())})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("x-csrf-token", xsrf)
	return request
}

func GenerateBotDislike(cookie string, id string) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://www.roblox.com/voting/vote?assetId=%s&vote=false", id), bytes.NewReader([]byte("")))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateBotFavorite(cookie string, id string) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://www.roblox.com/v2/favorite/toggle?itemTargetId=%s&favoriteType=Asset", id), bytes.NewReader([]byte("")))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateBotDisable(cookie string, id string) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	body := []byte("{\"name\":\"[ Content Deleted ]\",\"description\":\"[ Content Deleted ]\",\"enableComments\":false,\"genres\":[\"All\"],\"isCopyingAllowed\":false}")
	request, _ := http.NewRequest("PATCH", fmt.Sprintf("https://develop.roblox.com/v1/assets/%s", id), bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateThumbnailRequest(cookie string, id string, icon []byte) *http.Request {
	xsrf := <-FetchXsrf(cookie)
	body := fmt.Sprintf(FormData, Boundary, "png", "icon", icon, Boundary)
	request, _ := http.NewRequest("POST", fmt.Sprintf("https://publish.roblox.com/v1/plugins/%s/icon", id), bytes.NewReader([]byte(body)))
	request.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", Boundary))

	request.AddCookie(&http.Cookie{Name: "RBXEventTrackerV2", Value: "browserid=51723628002"})
	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	request.Header.Add("X-CSRF-TOKEN", xsrf)
	return request
}

func GenerateAnalytics1(cookie string, id string) *http.Request {
	request, _ := http.NewRequest("GET", fmt.Sprintf("http://ecsv2.roblox.com/pe?t=Studio&category=Action&evt=Insert&label=%s&value=0", id), bytes.NewReader([]byte("")))

	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	return request
}

func GenerateAnalytics2(cookie string, id string, name string) *http.Request {
	request, _ := http.NewRequest("GET", fmt.Sprintf("http://ecsv2.roblox.com/pe?t=Studio&ctx=click&evt=toolboxInsert&currentCategory=FreeModels&assetId=%s&searchText=%s&clientId=A&studioSid=A&placeId=0",
		id, gourl.QueryEscape(name)), bytes.NewReader([]byte("")))

	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	return request
}

func GenerateAnalytics3(cookie string, id string) *http.Request {
	request, _ := http.NewRequest("GET", fmt.Sprintf("http://ecsv2.roblox.com/pe?t=Studio&evt=InsertRemains120&label=%s&value=0", id), bytes.NewReader([]byte("")))

	request.AddCookie(&http.Cookie{Name: ".ROBLOSECURITY", Value: cookie})
	request.Header.Add("User-Agent", UserAgent)
	return request
}


func init() {
	ProxylessClient = &http.Client{}
}
