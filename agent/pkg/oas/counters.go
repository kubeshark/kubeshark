package oas

import (
	"fmt"
	"github.com/chanced/openapi"
	"math"
	"strings"
)

type Counter struct {
	Entries     int     `json:"entries"`
	Failures    int     `json:"failures"`
	FirstSeen   float64 `json:"firstSeen"`
	LastSeen    float64 `json:"lastSeen"`
	SumRT       float64 `json:"sumRT"`
	SumDuration float64 `json:"sumDuration"`
}

func (c *Counter) addEntry(ts float64, rt float64, succ bool, dur float64) {
	if dur < 0 {
		panic("Duration cannot be negative")
	}

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

func setCounterMsgIfOk(oldStr string, cnt *Counter) string {
	tpl := "Mizu observed %d entries (%d failed), at %.3f hits/s, average response time is %.3f seconds"
	if oldStr == "" || (strings.HasPrefix(oldStr, "Mizu ") && strings.HasSuffix(oldStr, " seconds")) {
		return fmt.Sprintf(tpl, cnt.Entries, cnt.Failures, cnt.SumDuration/float64(cnt.Entries), cnt.SumRT/float64(cnt.Entries))
	}
	return oldStr
}

type CounterMaps struct {
	counterTotal    Counter
	counterMapTotal CounterMap
}

func (m *CounterMaps) processOp(opObj *openapi.Operation) error {
	if _, ok := opObj.Extensions.Extension(CountersTotal); ok {
		counter := new(Counter)
		err := opObj.Extensions.DecodeExtension(CountersTotal, counter)
		if err != nil {
			return err
		}
		m.counterTotal.addOther(counter)

		opObj.Description = setCounterMsgIfOk(opObj.Description, counter)
	}

	if _, ok := opObj.Extensions.Extension(CountersPerSource); ok {
		counterMap := new(CounterMap)
		err := opObj.Extensions.DecodeExtension(CountersPerSource, counterMap)
		if err != nil {
			return err
		}
		m.counterMapTotal.addOther(counterMap)
	}
	return nil
}

func (m *CounterMaps) processOas(oas *openapi.OpenAPI) error {
	if oas.Extensions == nil {
		oas.Extensions = openapi.Extensions{}
	}

	err := oas.Extensions.SetExtension(CountersTotal, m.counterTotal)
	if err != nil {
		return err
	}

	err = oas.Extensions.SetExtension(CountersPerSource, m.counterMapTotal)
	if err != nil {
		return nil
	}
	return nil
}
