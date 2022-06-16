import React, {useEffect, useMemo, useState} from "react";
import {Backdrop, Box, Button, Fade, Modal} from "@mui/material";
import styles from "./TrafficStatsModal.module.sass";
import closeIcon from "assets/close.svg";
import {Cell, Legend, Pie, PieChart, Tooltip} from "recharts";
import {Utils} from "../../../helpers/Utils";
import {TrafficPieChart} from "./TrafficPieChart/TrafficPieChart";

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


const mock = [
  {
    name: "HTTP",
    reqCount: 400,
    byteCount: 1000,
    commands: [
      {
        name: "POST",
        reqCount: 150,
        byteCount: 400
      },
      {
        name: "GET",
        reqCount: 200,
        byteCount: 500
      },
      {
        name: "PUT",
        reqCount: 50,
        byteCount: 100
      }
    ]
  },
  {
    name: "KAFKA",
    reqCount: 100,
    byteCount: 300,
    commands: [
      {
        name: "COMMAND1",
        reqCount: 70,
        byteCount: 200
      },
      {
        name: "COMMAND2",
        reqCount: 30,
        byteCount: 100
      }
    ]
  }
]

enum StatsMode {
  REQUESTS = "entriesCount",
  VOLUME = "volumeSizeBytes"
}

interface TrafficStatsModalProps {
  isOpen: boolean;
  onClose: () => void;
  data: any; // todo: create model
}

export const TrafficStatsModal: React.FC<TrafficStatsModalProps> = ({ isOpen, onClose, data }) => {

  const modes = Object.keys(StatsMode).filter(x => !(parseInt(x) >= 0));
  const [statsMode, setStatsMode] = useState(modes[0]);

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
            <img src={closeIcon} alt="close" onClick={() => onClose()} style={{ cursor: "pointer", userSelect: "none" }}/>
          </div>
          <div className={styles.title}>Traffic Statistics</div>
          <div className={styles.mainContainer}>
            <div>
              <span style={{marginRight: 15}}>Breakdown By</span>
              <select className={styles.select} value={statsMode} onChange={(e) => setStatsMode(e.target.value)}>
                {modes.map(mode => <option value={mode}>{mode}</option>)}
              </select>
            </div>
           <TrafficPieChart pieChartMode={statsMode} data={data}/>
          </div>
        </Box>
      </Fade>
    </Modal>
  );
}
