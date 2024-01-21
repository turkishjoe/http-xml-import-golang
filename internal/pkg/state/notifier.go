package state

// В нашем случае это нужно для проверки совершается ли
// на данный момент импорт или нет(например, в методе /state).
type Notifier interface {
	Notify(value bool)
	ReadValue() bool
}
