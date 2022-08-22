package observer

import (
	"github.com/lyesteven/go-framework/common/utils/observer/event"
)

type Observer interface {
	OnNotify(event.Event)
}
