package notifier

import (
	"errors"
	"fmt"
	"github.com/lyesteven/go-framework/common/utils/observer/event"
	"github.com/lyesteven/go-framework/common/utils/observer/observer"
	"sync"
)

type Notifier interface {
	Register(observer observer.Observer, observerID string) error
	Deregister(observerID string) error
	Notify(event.Event)
}

type EventCenter struct {
	items map[string]observer.Observer
	lock  sync.Mutex
}

func NewEventCenter() *EventCenter {
	return &EventCenter{
		items: make(map[string]observer.Observer, 8),
	}
}

func (e *EventCenter) Register(observer observer.Observer, observerID string) error {
	if observerID == "" {
		return errors.New("observerID is null")
	}
	e.lock.Lock()
	defer e.lock.Unlock()

	if _, ok := e.items[observerID]; ok {
		return fmt.Errorf("observerID:%s exists", observerID)
	}
	e.items[observerID] = observer
	return nil
}

func (e *EventCenter) Deregister(observerID string) error {
	if observerID == "" {
		return errors.New("observerID is null")
	}
	e.lock.Lock()
	defer e.lock.Unlock()

	if _, ok := e.items[observerID]; !ok {
		return fmt.Errorf("observerID:%s not exist", observerID)
	}
	delete(e.items, observerID)
	return nil
}

func (e *EventCenter) Notify(event event.Event) {
	e.lock.Lock()
	defer e.lock.Unlock()

	for _, v := range e.items {
		v.OnNotify(event)
	}
}
