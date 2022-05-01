package tap

import (
	"sync"
	"time"

	"github.com/google/gopacket/reassembly"
	"github.com/up9inc/mizu/logger"
	"github.com/up9inc/mizu/tap/api"
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
	streamsMap        api.TcpStreamMap
}

func (cl *Cleaner) clean() {
	startCleanTime := time.Now()

	cl.assemblerMutex.Lock()
	logger.Log.Debugf("Assembler Stats before cleaning %s", cl.assembler.Dump())
	flushed, closed := cl.assembler.FlushCloseOlderThan(startCleanTime.Add(-cl.connectionTimeout))
	cl.assemblerMutex.Unlock()

	cl.streamsMap.Range(func(k, v interface{}) bool {
		reqResMatchers := v.(api.TcpStream).GetReqResMatchers()
		for _, reqResMatcher := range reqResMatchers {
			if reqResMatcher == nil {
				continue
			}
			deleted := deleteOlderThan(reqResMatcher.GetMap(), startCleanTime.Add(-cl.connectionTimeout))
			cl.stats.deleted += deleted
		}
		return true
	})

	cl.statsMutex.Lock()
	logger.Log.Debugf("Assembler Stats after cleaning %s", cl.assembler.Dump())
	cl.stats.flushed += flushed
	cl.stats.closed += closed
	cl.statsMutex.Unlock()
}

func (cl *Cleaner) start() {
	go func() {
		ticker := time.NewTicker(cl.cleanPeriod)

		for {
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

func deleteOlderThan(matcherMap *sync.Map, t time.Time) int {
	numDeleted := 0

	if matcherMap == nil {
		return numDeleted
	}

	matcherMap.Range(func(key interface{}, value interface{}) bool {
		message, _ := value.(*api.GenericMessage)
		// TODO: Investigate the reason why `request` is `nil` in some rare occasion
		if message != nil {
			if message.CaptureTime.Before(t) {
				matcherMap.Delete(key)
				numDeleted++
			}
		}
		return true
	})

	return numDeleted
}
