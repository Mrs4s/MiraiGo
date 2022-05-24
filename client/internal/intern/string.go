package intern

import (
	"sync"
)

// String Interning is a technique for reducing the memory footprint of large
// strings. It can re-use strings that are already in memory.

type StringInterner struct {
	mu      sync.RWMutex
	strings map[string]string
}

func NewStringInterner() *StringInterner {
	return &StringInterner{
		strings: make(map[string]string),
	}
}

func (i *StringInterner) Intern(s string) string {
	i.mu.RLock()
	if v, ok := i.strings[s]; ok {
		i.mu.RUnlock()
		return v
	}
	i.mu.RUnlock()
	i.mu.Lock()
	i.strings[s] = s
	i.mu.Unlock()
	return s
}
