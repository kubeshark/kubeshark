import React, { useState } from "react";
import { Button } from "@mui/material";
import Api, { KubesharkWebsocketURL } from "../../../helpers/api";
import debounce from 'lodash/debounce';
import { useRecoilState } from "recoil";
import { useCommonStyles } from "../../../helpers/commonStyle"
import serviceMapModalOpenAtom from "../../../recoil/serviceMapModalOpen";
import { TrafficViewer } from "../../TrafficViewer/TrafficViewer"
import "../../../index.sass"
import oasModalOpenAtom from "../../../recoil/oasModalOpen/atom";
import serviceMap from "./assets/serviceMap.svg";
import services from "./assets/services.svg";
import trafficStatsIcon from "./assets/trafficStats.svg";
import trafficStatsModalOpenAtom from "../../../recoil/trafficStatsModalOpen";
import { REPLAY_ENABLED } from "../../../consts";

const api = Api.getInstance();

export const TrafficPage: React.FC = () => {
  const commonClasses = useCommonStyles();
  const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);
  const [openOasModal, setOpenOasModal] = useRecoilState(oasModalOpenAtom);
  const [trafficStatsModalOpen, setTrafficStatsModalOpen] = useRecoilState(trafficStatsModalOpenAtom);
  const [shouldCloseWebSocket, setShouldCloseWebSocket] = useState(false);

  const trafficViewerApi = { ...api }

  const handleOpenOasModal = () => {
    setShouldCloseWebSocket(true)
    setOpenOasModal(true);
  }

  const handleOpenStatsModal = () => {
    setShouldCloseWebSocket(true)
    setTrafficStatsModalOpen(true);
  }

  const openServiceMapModalDebounce = debounce(() => {
    setShouldCloseWebSocket(true)
    setServiceMapModalOpen(true)
  }, 500);

  const actionButtons = <div style={{ display: 'flex', height: "100%" }}>
    {window["isOasEnabled"] && <Button
      startIcon={<img className="custom" src={services} alt="services" />}
      size="large"
      variant="contained"
      className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
      style={{ marginRight: 25, textTransform: 'unset' }}
      onClick={handleOpenOasModal}>
      Service Catalog
    </Button>}
    {window["isServiceMapEnabled"] && <Button
      startIcon={<img src={serviceMap} className="custom" alt="service-map" style={{ marginRight: "8%" }} />}
      size="large"
      variant="contained"
      className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
      onClick={openServiceMapModalDebounce}
      style={{ marginRight: 25, textTransform: 'unset' }}>
      Service Map
    </Button>}
    <Button
      startIcon={<img className="custom" src={trafficStatsIcon} alt="services" />}
      size="large"
      variant="contained"
      className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
      style={{ textTransform: 'unset' }}
      onClick={handleOpenStatsModal}>
      Traffic Stats
    </Button>
  </div>

  return (
    <>
      <TrafficViewer webSocketUrl={KubesharkWebsocketURL} shouldCloseWebSocket={shouldCloseWebSocket} setShouldCloseWebSocket={setShouldCloseWebSocket}
        trafficViewerApiProp={trafficViewerApi} actionButtons={actionButtons} isShowStatusBar={!(openOasModal || serviceMapModalOpen || trafficStatsModalOpen)} isDemoBannerView={false} entryDetailedConfig={{
          isReplayEnabled: REPLAY_ENABLED
        }} />
    </>
  );
};
