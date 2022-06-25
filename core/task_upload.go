package core

import (
	"fmt"
	. "kekops/ui"
	"strconv"
)

type UploadTask struct {
	BaseTask
}

func (b *UploadTask) Execute() {
	imp := b.TDestination.(*Import)
	data := imp.Data()
	if len(data) == 0 {
		return
	}

	for {
		cookie := BotHolder
		asset := Proxies[0].TryRequest(GenerateBotUpload(cookie, imp.ToolboxName, data, false, false))
		if asset, err := strconv.Atoi(asset); err != nil {
			Console.Log(CATEGORY_OPERATIONS, LOG_ERROR, err.Error())
			continue
		} else {
			Console.Log(CATEGORY_OPERATIONS, LOG_INFO, fmt.Sprintf("Asset ID = %d", asset))
			break
		}
	}
}

func (b *UploadTask) Color() string {
	return "yellow"
}

func (b *UploadTask) Type() string {
	return "UPLOAD"
}
