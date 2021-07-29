package utils

import "sync"

// UploadWaiter 用于控制并发上传，当有一个文件多次上传时，
// 等待第一个上传，后续的上传并发进行(可以秒传).
type UploadWaiter struct {
	mu sync.Mutex
	m  map[string]*sync.WaitGroup
}

// NewUploadWaiter return a new UploadWaiter.
func NewUploadWaiter() *UploadWaiter {
	return &UploadWaiter{
		m: make(map[string]*sync.WaitGroup),
	}
}

// Wait 如果不是第一个上传则等待。
func (s *UploadWaiter) Wait(key string) {
	s.mu.Lock()
	if w, ok := s.m[key]; ok {
		s.mu.Unlock()
		w.Wait()
	} else {
		wg := new(sync.WaitGroup)
		wg.Add(1)
		s.m[key] = wg
		s.mu.Unlock()
	}
}

// Done 当前上传任务已完成。
func (s *UploadWaiter) Done(key string) {
	s.mu.Lock()
	if w, ok := s.m[key]; ok {
		w.Done()
		delete(s.m, key)
	}
	s.mu.Unlock()
}
