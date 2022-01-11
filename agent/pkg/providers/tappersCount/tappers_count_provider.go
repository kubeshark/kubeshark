package tappersCount

import "sync"

var lock = &sync.Mutex{}

var tappersCount int

func Add() {
	lock.Lock()
	defer lock.Unlock()

	tappersCount++
}

func Remove() {
	lock.Lock()
	defer lock.Unlock()

	tappersCount--
}

func Get() int {
	return tappersCount
}
