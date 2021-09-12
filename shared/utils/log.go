package utils

import (
	"fmt"
	"sync"
	"github.com/romana/rlog"
)

var outputLevel int
var errorsMap map[string]uint
var errorsMapMutex sync.Mutex
var nErrors uint

const baseStreamChannelTimeoutMs int = 5000 * 100
 
func logError(minOutputLevel int, t string, s string, a ...interface{}) {
	errorsMapMutex.Lock()
	nErrors++
	nb, _ := errorsMap[t]
	errorsMap[t] = nb + 1
	errorsMapMutex.Unlock()

	if outputLevel >= minOutputLevel {
		formatStr := fmt.Sprintf("%s: %s", t, s)
		rlog.Errorf(formatStr, a...)
	}
}
func Error(t string, s string, a ...interface{}) {
	logError(0, t, s, a...)
}
func SilentError(t string, s string, a ...interface{}) {
	logError(2, t, s, a...)
}
func Debug(s string, a ...interface{}) {
	rlog.Debugf(s, a...)
}
func Trace(s string, a ...interface{}) {
	rlog.Tracef(1, s, a...)
}
