package debounce

import (
	"fmt"
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
	running  bool
	canceled bool
	timeout  time.Duration
	timer    *time.Timer
}

func (d *Debouncer) setTimeout(timeout time.Duration) {
	// TODO: Return err if d.running
	d.timeout = timeout
}

func (d *Debouncer) setCallback(callback func()) {
	callbackWrapped := func() {
		if !d.canceled {
			callback()
		}
		d.running = false
	}

	d.callback = callbackWrapped
}

func (d *Debouncer) Cancel() {
	d.canceled = true
}

func (d *Debouncer) SetOn() error {
	if d.canceled {
		return fmt.Errorf("debouncer cancelled")
	}
	if d.running {
		return nil
	}

	d.running = true
	d.timer = time.AfterFunc(d.timeout, d.callback)
	return nil
}

func (d *Debouncer) IsOn() bool {
	return d.running
}
