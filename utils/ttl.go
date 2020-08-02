package utils

import (
	"sync"
	"time"
)

type TTList struct {
	list []*item
	lock *sync.Mutex
}

type item struct {
	i          interface{}
	lastAccess int64
}

func NewTTList(ttl int64) *TTList {
	l := &TTList{
		lock: new(sync.Mutex),
	}
	go func() {
		for now := range time.Tick(time.Second * 5) {
			l.lock.Lock()
			pos := 0
			for _, i := range l.list {
				if now.Unix()-i.lastAccess > ttl {
					l.list = append(l.list[:pos], l.list[pos+1:]...)
					if pos > 0 {
						pos++
					}
				}
				pos++
			}
			l.lock.Unlock()
		}
	}()
	return l
}

func (l *TTList) Add(i interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.list = append(l.list, &item{
		i:          i,
		lastAccess: time.Now().Unix(),
	})
}

func (l *TTList) Any(filter func(i interface{}) bool) bool {
	l.lock.Lock()
	defer l.lock.Unlock()
	for _, it := range l.list {
		if filter(it.i) {
			it.lastAccess = time.Now().Unix()
			return true
		}
	}
	return false
}
