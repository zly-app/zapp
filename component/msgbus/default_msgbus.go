package msgbus

import (
	"github.com/zly-app/zapp/core"
)

var defMsgBus = NewMsgbus()

func GetMsgbus() core.IMsgbus { return defMsgBus }
