package state

import "sync"

type Notifier interface {
	Notify(value bool)
	ReadValue() bool
}

type locker struct {
	mutex sync.RWMutex
	value bool
}

func NewNotifier() Notifier {
	return &locker{}
}

func (locker *locker) Notify(value bool) {
	locker.mutex.Lock()
	defer locker.mutex.Unlock()
	locker.value = value
}

func (locker *locker) ReadValue() bool {
	locker.mutex.RLock()
	defer locker.mutex.RUnlock()
	return locker.value
}
