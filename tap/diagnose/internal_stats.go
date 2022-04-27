package diagnose

import "github.com/up9inc/mizu/logger"

type tapperInternalStats struct {
	Ipdefrag            int
	MissedBytes         int
	Pkt                 int
	Sz                  int
	Totalsz             int
	RejectFsm           int
	RejectOpt           int
	RejectConnFsm       int
	Reassembled         int
	OutOfOrderBytes     int
	OutOfOrderPackets   int
	BiggestChunkBytes   int
	BiggestChunkPackets int
	OverlapBytes        int
	OverlapPackets      int
}

var InternalStats *tapperInternalStats

func InitializeTapperInternalStats() {
	InternalStats = &tapperInternalStats{}
}

func (stats *tapperInternalStats) PrintStatsSummary() {
	logger.Log.Infof("IPdefrag:\t\t%d", stats.Ipdefrag)
	logger.Log.Infof("TCP stats:")
	logger.Log.Infof(" missed bytes:\t\t%d", stats.MissedBytes)
	logger.Log.Infof(" total packets:\t\t%d", stats.Pkt)
	logger.Log.Infof(" rejected FSM:\t\t%d", stats.RejectFsm)
	logger.Log.Infof(" rejected Options:\t%d", stats.RejectOpt)
	logger.Log.Infof(" reassembled bytes:\t%d", stats.Sz)
	logger.Log.Infof(" total TCP bytes:\t%d", stats.Totalsz)
	logger.Log.Infof(" conn rejected FSM:\t%d", stats.RejectConnFsm)
	logger.Log.Infof(" reassembled chunks:\t%d", stats.Reassembled)
	logger.Log.Infof(" out-of-order packets:\t%d", stats.OutOfOrderPackets)
	logger.Log.Infof(" out-of-order bytes:\t%d", stats.OutOfOrderBytes)
	logger.Log.Infof(" biggest-chunk packets:\t%d", stats.BiggestChunkPackets)
	logger.Log.Infof(" biggest-chunk bytes:\t%d", stats.BiggestChunkBytes)
	logger.Log.Infof(" overlap packets:\t%d", stats.OverlapPackets)
	logger.Log.Infof(" overlap bytes:\t\t%d", stats.OverlapBytes)
}
