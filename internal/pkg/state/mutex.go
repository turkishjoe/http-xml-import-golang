package state

import "sync"

// Данный класс реализует мютекс для чтения bool значения
// Можно расширить этот класс не только для bool
type mutexNotifier struct {
	mutex sync.RWMutex
	value bool
}

func NewNotifier() Notifier {
	return &mutexNotifier{}
}

func (locker *mutexNotifier) Notify(value bool) {
	locker.mutex.Lock()
	defer locker.mutex.Unlock()
	locker.value = value
}

func (locker *mutexNotifier) ReadValue() bool {
	locker.mutex.RLock()
	defer locker.mutex.RUnlock()
	return locker.value
}
