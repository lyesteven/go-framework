package set

import "sync"

var pool sync.Pool

type Set struct {
	items map[interface{}]struct{}
	mutex sync.RWMutex
}

func (set *Set) Add(items ...interface{}) {
	set.mutex.Lock()
	defer set.mutex.Unlock()

	for _, item := range items {
		set.items[item] = struct{}{}
	}
}

func (set *Set) Remove(items ...interface{}) {
	set.mutex.Lock()
	defer set.mutex.Unlock()

	for _, item := range items {
		delete(set.items, item)
	}
}

func (set *Set) Exists(item interface{}) bool {
	set.mutex.RLock()
	_, ok := set.items[item]
	set.mutex.RUnlock()
	return ok
}

func (set *Set) Len() int {
	set.mutex.RLock()
	size := len(set.items)
	set.mutex.RUnlock()
	return size
}

func (set *Set) Clear() {
	set.mutex.Lock()
	set.items = map[interface{}]struct{}{}
	set.mutex.Unlock()
}

func (set *Set) Dispose() {
	set.mutex.Lock()
	defer set.mutex.Unlock()

	for k := range set.items {
		delete(set.items, k)
	}

	pool.Put(set)
}

func New(items ...interface{}) *Set {
	set := pool.Get().(*Set)
	for _, item := range items {
		set.items[item] = struct{}{}
	}
	return set
}

func init() {
	pool.New = func() interface{} {
		return &Set{
			items:make(map[interface{}]struct{}, 8),
		}
	}
}