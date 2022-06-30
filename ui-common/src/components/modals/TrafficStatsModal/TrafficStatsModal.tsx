import React, { useCallback, useEffect, useState } from "react";
import { Backdrop, Box, Button, debounce, Fade, Modal } from "@mui/material";
import styles from "./TrafficStatsModal.module.sass";
import closeIcon from "assets/close.svg";
import { TrafficPieChart } from "./TrafficPieChart/TrafficPieChart";
import { TimelineBarChart } from "./TimelineBarChart/TimelineBarChart";
import spinnerImg from "assets/spinner.svg";
import refreshIcon from "assets/refresh.svg";
import { useCommonStyles } from "../../../helpers/commonStyle";

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

export const PROTOCOLS = ["ALL PROTOCOLS","gRPC", "REDIS", "HTTP", "GQL", "AMQP", "KFAKA"];
export const ALL_PROTOCOLS = PROTOCOLS[0];

export const TrafficStatsModal: React.FC<TrafficStatsModalProps> = ({ isOpen, onClose, getPieStatsDataApi, getTimelineStatsDataApi }) => {

  const modes = Object.keys(StatsMode).filter(x => !(parseInt(x) >= 0));
  const [statsMode, setStatsMode] = useState(modes[0]);
  const [selectedProtocol, setSelectedProtocol] = useState("ALL PROTOCOLS");
  const [pieStatsData, setPieStatsData] = useState(null);
  const [timelineStatsData, setTimelineStatsData] = useState(null);
  const [isLoading, setIsLoading] = useState(false);
  const commonClasses = useCommonStyles();

  const getTrafficStats = useCallback(async () => {
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

  useEffect(() => {
    getTrafficStats();
  }, [getTrafficStats])

  const refreshStats = debounce(() => {
    getTrafficStats();
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
            <Button style={{ marginLeft: "2%", textTransform: 'unset' }}
              startIcon={<img src={refreshIcon} className="custom" alt="refresh"></img>}
              size="medium"
              variant="contained"
              className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
              onClick={refreshStats}
            >
              Refresh
            </Button>
          </div>
          <div className={styles.mainContainer}>
            <div className={styles.selectContainer}>
              <div>
                <span style={{ marginRight: 15 }}>Breakdown By</span>
                <select className={styles.select} value={statsMode} onChange={(e) => setStatsMode(e.target.value)}>
                  {modes.map(mode => <option key={mode} value={mode}>{mode}</option>)}
                </select>
              </div>
              <div>
                <span style={{ marginRight: 15 }}>Protocol</span>
                <select className={styles.select} value={selectedProtocol} onChange={(e) => setSelectedProtocol(e.target.value)}>
                  {PROTOCOLS.map(protocol => <option key={protocol} value={protocol}>{protocol}</option>)}
                </select>
              </div>
            </div>
            <div>
              {isLoading ? <div style={{ textAlign: "center", marginTop: 20 }}>
                <img alt="spinner" src={spinnerImg} style={{ height: 50 }} />
              </div> :
                <div>
                  <TrafficPieChart pieChartMode={statsMode} data={pieStatsData} selectedProtocol={selectedProtocol}/>
                  <TimelineBarChart timeLineBarChartMode={statsMode} data={timelineStatsData} selectedProtocol={selectedProtocol}/>
                </div>}
            </div>
          </div>
        </Box>
      </Fade>
    </Modal>
  );
}
