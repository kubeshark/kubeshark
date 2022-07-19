import React, { useCallback, useEffect, useState } from "react";
import { Backdrop, Box, debounce, Fade, Modal } from "@mui/material";
import styles from "./TrafficStatsModal.module.sass";
import closeIcon from "assets/close.svg";
import { TrafficPieChart } from "./TrafficPieChart/TrafficPieChart";
import { TimelineBarChart } from "./TimelineBarChart/TimelineBarChart";
import { LoadingWrapper } from "../../UI/withLoading/withLoading";
import { ALL_PROTOCOLS, StatsMode } from "./consts";
import { TimeRangePicker } from "./TimelineBarChart/TimeRangePicker/TimeTangePicker";

const modalStyle = {
  position: 'absolute',
  top: '6%',
  left: '50%',
  transform: 'translate(-50%, 0%)',
  width: '60vw',
  height: '90vh',
  bgcolor: 'background.paper',
  borderRadius: '5px',
  boxShadow: 24,
  p: 4,
  color: '#000',
};

interface TrafficStatsModalProps {
  isOpen: boolean;
  onClose: () => void;
  getTrafficStatsDataApi: (start?, end?) => Promise<any>
}

export const TrafficStatsModal: React.FC<TrafficStatsModalProps> = ({ isOpen, onClose, getTrafficStatsDataApi }) => {
  const modes = Object.keys(StatsMode).filter(x => !(parseInt(x) >= 0));
  const [statsMode, setStatsMode] = useState(modes[0]);
  const [selectedProtocol, setSelectedProtocol] = useState(ALL_PROTOCOLS);
  const [pieStatsData, setPieStatsData] = useState(null);
  const [timelineStatsData, setTimelineStatsData] = useState(null);
  const [protocols, setProtocols] = useState([])
  const [isLoading, setIsLoading] = useState(false);

  const getTrafficStats = useCallback(async (startTime, endTime) => {
    if (isOpen && getTrafficStatsDataApi) {
      (async () => {
        try {
          setIsLoading(true);
          const statsData = await getTrafficStatsDataApi(startTime, endTime);
          setPieStatsData(statsData.pie);
          setTimelineStatsData(statsData.timeline);
          setProtocols(statsData.protocols)
        } catch (e) {
          console.error(e)
        } finally {
          setIsLoading(false)
        }
      })()
    }
  }, [isOpen, getTrafficStatsDataApi, setPieStatsData, setTimelineStatsData])

  useEffect(() => {
    const now = new Date().getTime();
    const halfAnHourAgo = now - (30 * 60 * 1000);
    getTrafficStats(halfAnHourAgo, now);
  }, [getTrafficStats])

  const refreshStats = debounce((newStartTime, newEndTime) => {
    getTrafficStats(newStartTime, newEndTime);
  }, 500);

  return (
    <Modal
      aria-labelledby="transition-modal-title"
      aria-describedby="transition-modal-description"
      open={isOpen}
      onClose={onClose}
      closeAfterTransition
      BackdropComponent={Backdrop}
      BackdropProps={{ timeout: 500 }}>
      <Fade in={isOpen}>
        <Box sx={modalStyle}>
          <div className={styles.closeIcon}>
            <img src={closeIcon} alt="close" onClick={() => onClose()} style={{ cursor: "pointer", userSelect: "none" }} />
          </div>
          <div className={styles.headlineContainer}>
            <div className={styles.title}>Traffic Statistics</div>
          </div>
          <div className={styles.mainContainer}>
            <div className={styles.selectContainer}>
              <div>
                <span className={styles.selectTitle}>Breakdown By</span>
                <select className={styles.select} value={statsMode} onChange={(e) => setStatsMode(e.target.value)}>
                  {modes.map(mode => <option key={mode} value={mode}>{mode}</option>)}
                </select>
              </div>
              <div>
                <span className={styles.selectTitle}>Protocol</span>
                <select className={styles.select} value={selectedProtocol} onChange={(e) => setSelectedProtocol(e.target.value)}>
                  {protocols.map(protocol => <option key={protocol} value={protocol}>{protocol}</option>)}
                </select>
              </div>
              <div>
                <TimeRangePicker refreshStats={refreshStats} />
              </div>
            </div>
            <div>
              <LoadingWrapper isLoading={isLoading} loaderMargin={20} loaderHeight={50}>
                <div>
                  <TrafficPieChart pieChartMode={statsMode} data={pieStatsData} selectedProtocol={selectedProtocol} />
                  <TimelineBarChart timeLineBarChartMode={statsMode} data={timelineStatsData} selectedProtocol={selectedProtocol} />
                </div>
              </LoadingWrapper>
            </div>
          </div>
        </Box>
      </Fade>
    </Modal>
  );
}
