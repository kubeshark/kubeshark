package oas

import "math"

type Counter struct {
	Entries   int     `json:"entries"`
	Failures  int     `json:"failures"`
	FirstSeen float64 `json:"firstSeen"`
	LastSeen  float64 `json:"lastSeen"`
	SumRT     float64 `json:"sumRT"`
}

func (c *Counter) addEntry(ts float64, rt float64, succ bool) {
	c.Entries += 1
	c.SumRT += rt
	if !succ {
		c.Failures += 1
	}

	if c.FirstSeen == 0 {
		c.FirstSeen = ts
	} else {
		c.FirstSeen = math.Min(c.FirstSeen, ts)
	}

	c.LastSeen = math.Max(c.LastSeen, ts)
}

// TODO: addOther

type CounterMap = map[string]*Counter
