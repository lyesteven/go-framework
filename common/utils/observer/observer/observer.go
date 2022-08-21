package observer

import (
	"gworld/git/GoFrameWork/common/utils/observer/event"
)

type Observer interface {
	OnNotify(event.Event)
}
