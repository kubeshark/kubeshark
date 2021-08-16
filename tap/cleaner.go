package tap

import (
	"sync"
	"time"

	"github.com/romana/rlog"

	"github.com/google/gopacket/reassembly"
)

type CleanerStats struct {
	flushed int
	closed  int
	deleted int
}

type Cleaner struct {
	assembler         *reassembly.Assembler
	assemblerMutex    *sync.Mutex
	cleanPeriod       time.Duration
	connectionTimeout time.Duration
	stats             CleanerStats
	statsMutex        sync.Mutex
}

func (cl *Cleaner) clean() {
	startCleanTime := time.Now()

	cl.assemblerMutex.Lock()
	rlog.Debugf("Assembler Stats before cleaning %s", cl.assembler.Dump())
	flushed, closed := cl.assembler.FlushCloseOlderThan(startCleanTime.Add(-cl.connectionTimeout))
	cl.assemblerMutex.Unlock()

	cl.statsMutex.Lock()
	rlog.Debugf("Assembler Stats after cleaning %s", cl.assembler.Dump())
	cl.stats.flushed += flushed
	cl.stats.closed += closed
	cl.statsMutex.Unlock()
}

func (cl *Cleaner) start() {
	go func() {
		ticker := time.NewTicker(cl.cleanPeriod)

		for true {
			<-ticker.C
			cl.clean()
		}
	}()
}

func (cl *Cleaner) dumpStats() CleanerStats {
	cl.statsMutex.Lock()

	stats := CleanerStats{
		flushed: cl.stats.flushed,
		closed:  cl.stats.closed,
		deleted: cl.stats.deleted,
	}

	cl.stats.flushed = 0
	cl.stats.closed = 0
	cl.stats.deleted = 0

	cl.statsMutex.Unlock()

	return stats
}
