package debounce

import (
	"time"
)

func NewDebouncer(timeout time.Duration, callback func()) *Debouncer {
	var debouncer Debouncer
	debouncer.setTimeout(timeout)
	debouncer.setCallback(callback)
	return &debouncer
}

type Debouncer struct {
	callback func()
	running bool
	timeout time.Duration
	timer *time.Timer
}

func (d *Debouncer) setTimeout(timeout time.Duration) {
	// TODO: Return err if d.running
	d.timeout = timeout
}

func (d *Debouncer) setCallback(callback func()) {
	callbackWrapped := func() {
		callback()
		d.running = false
	}

	d.callback = callbackWrapped
}

func (d *Debouncer) SetOn() {
	if d.running == true {
		return
	}

	d.running = true
	d.timer = time.AfterFunc(d.timeout, d.callback)
}
