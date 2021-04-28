package main

import (
	"sync"
)

type AppStats struct {
	matchedMessages int
}

type StatsTracker struct {
	stats             AppStats
	statsMutex	  sync.Mutex
}

func (st *StatsTracker) incMatchedMessages() {
	st.statsMutex.Lock()
	st.stats.matchedMessages++
	st.statsMutex.Unlock()
}

func (st *StatsTracker) dumpStats() AppStats {
	st.statsMutex.Lock()

	stats := AppStats{
		matchedMessages: st.stats.matchedMessages,
	}

	st.stats.matchedMessages = 0

	st.statsMutex.Unlock()

	return stats
}

