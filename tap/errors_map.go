package tap

import (
	"fmt"
	"sync"

	"github.com/romana/rlog"
)

type errorsMap struct {
	errorsMap      map[string]int
	outputLevel    int
	nErrors        int
	errorsMapMutex sync.Mutex
}

func NewErrorsMap(outputLevel int) *errorsMap {
	return &errorsMap{
		errorsMap:   make(map[string]int),
		outputLevel: outputLevel,
	}
}

/* minOutputLevel: Error will be printed only if outputLevel is above this value
 * t:              key for errorsMap (counting errors)
 * s, a:           arguments log.Printf
 * Note:           Too bad for perf that a... is evaluated
 */
func (e *errorsMap) logError(minOutputLevel int, t string, s string, a ...interface{}) {
	e.errorsMapMutex.Lock()
	e.nErrors++
	nb := e.errorsMap[t]
	e.errorsMap[t] = nb + 1
	e.errorsMapMutex.Unlock()

	if e.outputLevel >= minOutputLevel {
		formatStr := fmt.Sprintf("%s: %s", t, s)
		rlog.Errorf(formatStr, a...)
	}
}

func (e *errorsMap) Error(t string, s string, a ...interface{}) {
	e.logError(0, t, s, a...)
}

func (e *errorsMap) SilentError(t string, s string, a ...interface{}) {
	e.logError(2, t, s, a...)
}

func (e *errorsMap) Debug(s string, a ...interface{}) {
	rlog.Debugf(s, a...)
}

func (e *errorsMap) Trace(s string, a ...interface{}) {
	rlog.Tracef(1, s, a...)
}

func (e *errorsMap) getErrorsSummary() (int, string) {
	e.errorsMapMutex.Lock()
	errorMapLen := len(e.errorsMap)
	errorsSummery := fmt.Sprintf("%v", e.errorsMap)
	e.errorsMapMutex.Unlock()
	return errorMapLen, errorsSummery
}
