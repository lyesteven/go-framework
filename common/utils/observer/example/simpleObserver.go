package main

import (
	"fmt"
	"github.com/lyesteven/go-framework/common/utils/observer/event"
	"github.com/lyesteven/go-framework/common/utils/observer/notifier"
)

type simpleEvent struct{}

// implement Event interface
func (s *simpleEvent) Echo() string {
	return "this is a simple event"
}

type simpleObserver struct {
	id string
}

// implement Observer interface
func (s *simpleObserver) OnNotify(event event.Event) {
	switch e := event.(type) {
	case *simpleEvent:
		fmt.Printf("%s: recv event[%s]\n", s.id, e.Echo())
	}
}

func main() {
	so1 := &simpleObserver{id: "observer-1"}
	so2 := &simpleObserver{id: "observer-2"}

	ec := notifier.NewEventCenter()
	ec.Register(so1, so1.id)
	ec.Register(so2, so2.id)
	defer ec.Deregister(so1.id)
	defer ec.Deregister(so2.id)

	e := &simpleEvent{}
	ec.Notify(e)
}
