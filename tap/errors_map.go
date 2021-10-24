package tap

import (
	"fmt"
	"sync"

	"github.com/up9inc/mizu/shared/logger"
)

type errorsMap struct {
	errorsMap      map[string]uint
	outputLevel    int
	nErrors        uint
	errorsMapMutex sync.Mutex
}

func NewErrorsMap(outputLevel int) *errorsMap {
	return &errorsMap{
		errorsMap:   make(map[string]uint),
		outputLevel: outputLevel,
	}
}

/* minOutputLevel: Error will be printed only if outputLevel is above this value
 * t:              key for errorsMap (counting errors)
 * s, a:           arguments logger.Log.Infof
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
		logger.Log.Errorf(formatStr, a...)
	}
}

func (e *errorsMap) Error(t string, s string, a ...interface{}) {
	e.logError(0, t, s, a...)
}

func (e *errorsMap) SilentError(t string, s string, a ...interface{}) {
	e.logError(2, t, s, a...)
}

func (e *errorsMap) Debug(s string, a ...interface{}) {
	logger.Log.Debugf(s, a...)
}

func (e *errorsMap) getErrorsSummary() (int, string) {
	e.errorsMapMutex.Lock()
	errorMapLen := len(e.errorsMap)
	errorsSummery := fmt.Sprintf("%v", e.errorsMap)
	e.errorsMapMutex.Unlock()
	return errorMapLen, errorsSummery
}

func (e *errorsMap) PrintSummary() {
	logger.Log.Infof("Errors: %d", e.nErrors)
	for t := range e.errorsMap {
		logger.Log.Infof(" %s:\t\t%d", e, e.errorsMap[t])
	}
}
