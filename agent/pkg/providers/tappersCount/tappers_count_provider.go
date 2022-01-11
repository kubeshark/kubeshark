package tappersCount

import "sync"

var lock = &sync.Mutex{}

var tappersCount int

func TapperAdded() {
	lock.Lock()
	defer lock.Unlock()

	tappersCount++
}

func TapperRemoved() {
	lock.Lock()
	defer lock.Unlock()

	tappersCount--
}

func Get() int {
	return tappersCount
}
