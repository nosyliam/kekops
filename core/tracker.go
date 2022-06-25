package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	. "kekops/ui"
)

var Tracker *ImportTracker

type Require struct {
	ID     string `json:"id"`
	Cookie string `json:"account"`
	User   string `json:"user"`
	Pass   string `json:"pass"`
}

type Group struct {
	Original string `json:"string"`
	ID       string `json:"id"`
}

type Import struct {
	ID          string      `json:"id"` // Filename
	Type        ImportType  `json:"type"`
	Asset       int         `json:"asset"` // Asset ID
	Original    string      `json:"original"`
	Mode        ImportLevel `json:"mode"`
	Name        string      `json:"name"`
	ToolboxName string      `json:"tbname"`
	LastUpdated int         `json:"lastUpdated"`
}

func (i *Import) Data() []byte {
	file, err := ioutil.ReadFile(fmt.Sprintf("cache/%s.rbxm", i.ID))
	if err != nil {
		Console.Log(CATEGORY_TRACKER, LOG_ERROR, fmt.Sprintf("Attempted read of unstored import %s", i.ID))
		return []byte("")
	}
	return file
}

func (i *Import) Icon() []byte {
	file, err := ioutil.ReadFile(fmt.Sprintf("cache/%s.png", i.ID))
	if err != nil {
		Console.Log(CATEGORY_TRACKER, LOG_ERROR, fmt.Sprintf("Attempted read of unstored import thumbnail %s", i.ID))
		return []byte("")
	}
	return file
}

func (i *Import) TimeElapsed() int {
	return int(time.Now().Unix()) - i.LastUpdated
}

type ImportTracker struct {
	Imports  []Import   `json:"imports"`
	Requires []Require  `json:"requires"`
	Groups   []Group    `json:"groups"`
	Models   []int      `json:"models"`
	File     *os.File   `json:"-"`
	Lock     sync.Mutex `json:"-"`
}

func (i *ImportTracker) AddGroup(original string, id string) {
	i.Groups = append(i.Groups, Group{Original: original, ID: id})
	i.Save()
}

func (i *ImportTracker) FindGroup(original string) *Group {
	for _, group := range i.Groups {
		if group.Original == original {
			return &group
		}
	}
	return nil
}

func (i *ImportTracker) AddRequire(id string, cookie *Account) {
	i.Requires = append([]Require{Require{id, cookie.cookie, cookie.user, cookie.pass}}, i.Requires...)
	i.Save()
}

func (i *ImportTracker) AddModel(id int) {
	i.Models = append(i.Models, id)
	i.Save()
}


func (i *ImportTracker) List() {
	out := "[Import List](fg:yellow)\n"
	for _, imp := range i.Imports {
		var mode string
		switch imp.Mode {
		case IMPORT_LEVEL_TOOLBOX:
			mode = "Toolbox"
		case IMPORT_LEVEL_FRONTPAGE:
			mode = "FrontPage"
		case IMPORT_LEVEL_BOTH:
			mode = "Double"
		default:
			mode = "Unmaintained"
		}
		out += fmt.Sprintf("%s - Asset: %d - Mode: %s\n", imp.ID, imp.Asset, mode)
	}
	Console.Log(CATEGORY_TRACKER, LOG_INFO, out)
}

func (i *ImportTracker) ImportModel(data []byte, id string, name string, tbName string) *Import {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if file, err := os.OpenFile(fmt.Sprintf("cache/%s.rbxm", id), os.O_RDWR|os.O_CREATE, 0755); err != nil {
		Console.Log(CATEGORY_TRACKER, LOG_ERROR, fmt.Sprintf("Failed to store import: %v", err))
		return nil
	} else {
		Console.Log(CATEGORY_TRACKER, LOG_INFO, fmt.Sprintf("Saving import %s...", id))
		file.Truncate(0)
		file.Write(data)
		file.Sync()
		file.Close()
	}

	iImport := Import{
		ID:          id,
		Type:        IMPORT_MODEL,
		Name:        name,
		ToolboxName: tbName,
		LastUpdated: int(time.Now().Unix()),
	}
	Tracker.Imports = append(Tracker.Imports, iImport)
	i.Save()
	return &iImport
}

func (i *ImportTracker) ImportPlugin(data []byte, icon []byte, id string, original string) *Import {
	i.Lock.Lock()
	defer i.Lock.Unlock()
	if file, err := os.OpenFile(fmt.Sprintf("cache/%s.rbxm", id), os.O_RDWR|os.O_CREATE, 0755); err != nil {
		Console.Log(CATEGORY_TRACKER, LOG_ERROR, fmt.Sprintf("Failed to store plugin import data: %v", err))
		return nil
	} else {
		Console.Log(CATEGORY_TRACKER, LOG_INFO, fmt.Sprintf("Saving import %s...", id))
		file.Truncate(0)
		file.Write(data)
		file.Sync()
		file.Close()
	}
	if file, err := os.OpenFile(fmt.Sprintf("cache/%s.png", id), os.O_RDWR|os.O_CREATE, 0755); err != nil {
		Console.Log(CATEGORY_TRACKER, LOG_ERROR, fmt.Sprintf("Failed to store plugin import data: %v", err))
		return nil
	} else {
		Console.Log(CATEGORY_TRACKER, LOG_INFO, fmt.Sprintf("Saving import icon..."))
		file.Truncate(0)
		file.Write(icon)
		file.Sync()
		file.Close()
	}

	iImport := Import{
		ID:          id,
		Type:        IMPORT_PLUGIN,
		Name:        id,
		Original:    original,
		LastUpdated: int(time.Now().Unix()),
	}
	Tracker.Imports = append(Tracker.Imports, iImport)
	i.Save()
	return &iImport
}

func (i *ImportTracker) Save() {
	if data, err := json.Marshal(i); err != nil {
		Console.Log(CATEGORY_TRACKER, LOG_ERROR, "Failed to save tracker to disk.")
	} else {
		i.File.Truncate(0)
		i.File.Seek(0, 0)
		i.File.Write(data)
		i.File.Sync()
	}
}

func (i *ImportTracker) UpdateAsset(id string, asset int) {
	for n, imp := range i.Imports {
		if imp.ID == id {
			i.Imports[n].Asset = asset
			Console.Log(CATEGORY_TRACKER, LOG_INFO, fmt.Sprintf("Successfully updated import %s", id))
		}
	}
	i.Save()
}

func (i *ImportTracker) UpdateMode(id string, mode ImportLevel) {
	for _, imp := range i.Imports {
		if imp.ID == id {
			if (imp.Mode == IMPORT_LEVEL_TOOLBOX && mode == IMPORT_LEVEL_FRONTPAGE) ||
				(imp.Mode == IMPORT_LEVEL_FRONTPAGE && mode == IMPORT_LEVEL_TOOLBOX) {
				imp.Mode = IMPORT_LEVEL_BOTH
			} else {
				imp.Mode = mode
			}
			Console.Log(CATEGORY_TRACKER, LOG_INFO, fmt.Sprintf("Successfully updated import %s", id))
		}
	}
	i.Save()
}

func (i *ImportTracker) UpdateName(id string, name string) {
	for n, imp := range i.Imports {
		if imp.ID == id {
			i.Imports[n].Name = name
			Console.Log(CATEGORY_TRACKER, LOG_INFO, fmt.Sprintf("Successfully updated import %s", id))
		}
	}
	i.Save()
}

func (i *ImportTracker) Remove(id string) bool {
	for n, imp := range i.Imports {
		if imp.ID == id {
			os.Remove(fmt.Sprintf("cache/%s.rbxm", imp.ID))
			Console.Log(CATEGORY_TRACKER, LOG_INFO, fmt.Sprintf("Import %s removed.", imp.ID))
			i.Imports = append(i.Imports[:n], i.Imports[n+1:]...)
			i.Save()
			return true
		}
	}
	return false
}

func (i *ImportTracker) Find(id string) *Import {
	for _, imp := range i.Imports {
		if imp.ID == id {
			return &imp
		}
	}
	return nil
}

func (i *ImportTracker) MarkTime(id string) {
	for n, imp := range i.Imports {
		if imp.ID == id {
			i.Imports[n].LastUpdated = int(time.Now().Unix())
		}
	}
	i.Save()
}

func (i *ImportTracker) Close() {
	i.Save()
	i.File.Close()
}

func (i *ImportTracker) ListRequires() {
	var list []string
	for _, imp := range i.Requires {
		list = append(list,imp.ID)
	}
	ioutil.WriteFile("list.txt", []byte(strings.Join(list, ",")), os.FileMode(077))
}

func init() {
	var err error
	Tracker = &ImportTracker{}
	os.Mkdir("cache", 0755)
	if Tracker.File, err = os.OpenFile("tracker.json", os.O_RDWR|os.O_CREATE, 0755); err != nil {
		panic(err)
	} else {
		data, _ := ioutil.ReadFile("tracker.json")
		data = []byte(strings.Replace(string(data), " ", "", -1))
		if err := json.Unmarshal(data, Tracker); err != nil {
			go func() {
				for {
					if Console != nil {
						break
					}
				}
				Console.Log(CATEGORY_TRACKER, LOG_INFO, "Successfully loaded tracker.")
				Tracker.Save()
			}()
		}
	}
}
