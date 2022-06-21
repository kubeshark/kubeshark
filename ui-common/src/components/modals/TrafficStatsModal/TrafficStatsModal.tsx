import React, { useEffect, useState } from "react";
import { Backdrop, Box, Fade, Modal } from "@mui/material";
import styles from "./TrafficStatsModal.module.sass";
import closeIcon from "assets/close.svg";
import { TrafficPieChart } from "./TrafficPieChart/TrafficPieChart";
import { TimelineBarChart } from "./TimelineBarChart/TimelineBarChart";
import spinnerImg from "assets/spinner.svg";

const modalStyle = {
  position: 'absolute',
  top: '6%',
  left: '50%',
  transform: 'translate(-50%, 0%)',
  width: '50vw',
  height: '82vh',
  bgcolor: 'background.paper',
  borderRadius: '5px',
  boxShadow: 24,
  p: 4,
  color: '#000',
};

export enum StatsMode {
  REQUESTS = "entriesCount",
  VOLUME = "volumeSizeBytes"
}

interface TrafficStatsModalProps {
  isOpen: boolean;
  onClose: () => void;
  getPieStatsDataApi: () => Promise<any>
  getTimelineStatsDataApi: () => Promise<any>
}

export const TrafficStatsModal: React.FC<TrafficStatsModalProps> = ({ isOpen, onClose, getPieStatsDataApi, getTimelineStatsDataApi }) => {

  const modes = Object.keys(StatsMode).filter(x => !(parseInt(x) >= 0));
  const [statsMode, setStatsMode] = useState(modes[0]);
  const [pieStatsData, setPieStatsData] = useState(null);
  const [timelineStatsData, setTimelineStatsData] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  useEffect(() => {
    if (isOpen && getPieStatsDataApi) {
      (async () => {
        try {
          setIsLoading(true);
          const pieData = await getPieStatsDataApi();
          setPieStatsData(pieData);
          const timelineData = await getTimelineStatsDataApi();
          setTimelineStatsData(timelineData);
        } catch (e) {
          console.error(e)
        } finally {
          setIsLoading(false)
        }
      })()
    }
  }, [isOpen, getPieStatsDataApi, getTimelineStatsDataApi, setPieStatsData, setTimelineStatsData])

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
          <div className={styles.title}>Traffic Statistics</div>
          <div className={styles.mainContainer}>
            <div>
              <span style={{ marginRight: 15 }}>Breakdown By</span>
              <select className={styles.select} value={statsMode} onChange={(e) => setStatsMode(e.target.value)}>
                {modes.map(mode => <option key={mode} value={mode}>{mode}</option>)}
              </select>
            </div>
            <div>
              {isLoading ? <div style={{ textAlign: "center", marginTop: 20 }}>
                <img alt="spinner" src={spinnerImg} style={{ height: 50 }} />
              </div> : 
                <div>
                  <TrafficPieChart pieChartMode={statsMode} data={pieStatsData} />
                  <TimelineBarChart timeLineBarChartMode={statsMode} data={timelineStatsData} />
                </div>}
            </div>
          </div>
        </Box>
      </Fade>
    </Modal>
  );
}
