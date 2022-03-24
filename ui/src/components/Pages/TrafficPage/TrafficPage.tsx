import React, {useEffect, useState} from "react";
import { Button } from "@material-ui/core";
import Api,{getWebsocketUrl} from "../../../helpers/api";
import debounce from 'lodash/debounce';
import {useSetRecoilState, useRecoilState} from "recoil";
import {useCommonStyles} from "../../../helpers/commonStyle"
import serviceMapModalOpenAtom from "../../../recoil/serviceMapModalOpen";
import TrafficViewer from "@up9/mizu-common"
import "@up9/mizu-common/dist/index.css"
import oasModalOpenAtom from "../../../recoil/oasModalOpen/atom";
import serviceMap from "../../assets/serviceMap.svg";
import services from "../../assets/services.svg";

interface TrafficPageProps {
  setAnalyzeStatus?: (status: any) => void;
}

const api = Api.getInstance();

export const TrafficPage: React.FC<TrafficPageProps> = ({setAnalyzeStatus}) => {
  const commonClasses = useCommonStyles();
  const setServiceMapModalOpen = useSetRecoilState(serviceMapModalOpenAtom);
  const [openOasModal, setOpenOasModal] = useRecoilState(oasModalOpenAtom);
  const [openWebSocket, setOpenWebSocket] = useState(true);

const trafficViewerApi = {...api}

  const handleOpenOasModal = () => {
    setOpenWebSocket(false)
    setOpenOasModal(true);
  }

  const openServiceMapModalDebounce = debounce(() => {
    setOpenWebSocket(false)
    setServiceMapModalOpen(true)
  }, 500);

  const actionButtons = (window["isOasEnabled"] || window["isServiceMapEnabled"]) &&
                          <div style={{ display: 'flex', height: "100%" }}>
                              {window["isOasEnabled"] && <Button
                                startIcon={<img className="custom" src={services} alt="services"></img>}
                                size="large"
                                type="submit"
                                variant="contained"
                                className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
                                style={{ marginRight: 25, textTransform: 'unset' }}
                                onClick={handleOpenOasModal}>
                                OpenApi Specs
                              </Button>}
                              {window["isServiceMapEnabled"] && <Button
                                startIcon={<img src={serviceMap} className="custom" alt="service-map" style={{marginRight:"8%"}}></img>}
                                size="large"
                                variant="contained"
                                className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
                                onClick={openServiceMapModalDebounce}
                                style={{textTransform: 'unset'}}>
                                Service Map
                              </Button>}
                        </div>

  useEffect(() => {
    return () => {
      //closeSocket()
    }
  },[])

  return (
  <>
      <TrafficViewer setAnalyzeStatus={setAnalyzeStatus} webSocketUrl={getWebsocketUrl()} isCloseWebSocket={!openWebSocket}
                     trafficViewerApiProp={trafficViewerApi} actionButtons={actionButtons} isShowStatusBar={!openOasModal} isDemoBannerView={false}/>
  </>
  );
};
