package tap

import (
	"sync"
	"time"

	"github.com/google/gopacket/reassembly"
	"github.com/kubeshark/kubeshark/logger"
	"github.com/kubeshark/kubeshark/tap/api"
)

type CleanerStats struct {
	deleted int
}

type Cleaner struct {
	assembler         *reassembly.Assembler
	cleanPeriod       time.Duration
	connectionTimeout time.Duration
	stats             CleanerStats
	statsMutex        sync.Mutex
	streamsMap        api.TcpStreamMap
}

func (cl *Cleaner) clean() {
	startCleanTime := time.Now()

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
		deleted: cl.stats.deleted,
	}

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
