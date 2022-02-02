package oas

import "math"

type Counter struct {
	Entries     int     `json:"entries"`
	Failures    int     `json:"failures"`
	FirstSeen   float64 `json:"firstSeen"`
	LastSeen    float64 `json:"lastSeen"`
	SumRT       float64 `json:"sumRT"`
	SumDuration float64 `json:"sumDuration"`
}

func (c *Counter) addEntry(ts float64, rt float64, succ bool, dur float64) {
	c.Entries += 1
	c.SumRT += rt
	c.SumDuration += dur
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

func (c *Counter) addOther(other *Counter) {
	c.Entries += other.Entries
	c.SumRT += other.SumRT
	c.Failures += other.Failures
	c.SumDuration += other.SumDuration

	if c.FirstSeen == 0 {
		c.FirstSeen = other.FirstSeen
	} else {
		c.FirstSeen = math.Min(c.FirstSeen, other.FirstSeen)
	}

	c.LastSeen = math.Max(c.LastSeen, other.LastSeen)
}

type CounterMap map[string]*Counter

func (m *CounterMap) addOther(other *CounterMap) {
	for src, cmap := range *other {
		if existing, ok := (*m)[src]; ok {
			existing.addOther(cmap)
		} else {
			copied := *cmap
			(*m)[src] = &copied
		}
	}
}
