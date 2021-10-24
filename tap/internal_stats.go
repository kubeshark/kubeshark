package tap

import "github.com/up9inc/mizu/shared/logger"

type tapperInternalStats struct {
	ipdefrag            int
	missedBytes         int
	pkt                 int
	sz                  int
	totalsz             int
	rejectFsm           int
	rejectOpt           int
	rejectConnFsm       int
	reassembled         int
	outOfOrderBytes     int
	outOfOrderPackets   int
	biggestChunkBytes   int
	biggestChunkPackets int
	overlapBytes        int
	overlapPackets      int
}

func NewTapperInternalStats() *tapperInternalStats {
	return &tapperInternalStats{}
}

func (stats *tapperInternalStats) PrintStatsSummary() {
	logger.Log.Infof("IPdefrag:\t\t%d", stats.ipdefrag)
	logger.Log.Infof("TCP stats:")
	logger.Log.Infof(" missed bytes:\t\t%d", stats.missedBytes)
	logger.Log.Infof(" total packets:\t\t%d", stats.pkt)
	logger.Log.Infof(" rejected FSM:\t\t%d", stats.rejectFsm)
	logger.Log.Infof(" rejected Options:\t%d", stats.rejectOpt)
	logger.Log.Infof(" reassembled bytes:\t%d", stats.sz)
	logger.Log.Infof(" total TCP bytes:\t%d", stats.totalsz)
	logger.Log.Infof(" conn rejected FSM:\t%d", stats.rejectConnFsm)
	logger.Log.Infof(" reassembled chunks:\t%d", stats.reassembled)
	logger.Log.Infof(" out-of-order packets:\t%d", stats.outOfOrderPackets)
	logger.Log.Infof(" out-of-order bytes:\t%d", stats.outOfOrderBytes)
	logger.Log.Infof(" biggest-chunk packets:\t%d", stats.biggestChunkPackets)
	logger.Log.Infof(" biggest-chunk bytes:\t%d", stats.biggestChunkBytes)
	logger.Log.Infof(" overlap packets:\t%d", stats.overlapPackets)
	logger.Log.Infof(" overlap bytes:\t\t%d", stats.overlapBytes)
}
