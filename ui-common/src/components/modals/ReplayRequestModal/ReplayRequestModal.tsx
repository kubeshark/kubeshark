import { Backdrop, Box, Button, Fade, Modal } from "@mui/material";
import React, { useCallback, useEffect, useState } from "react";
import { useCommonStyles } from "../../../helpers/commonStyle";
import { Tabs } from "../../UI";
import { SectionsRepresentation } from "../../EntryDetailed/EntryViewer/EntryViewer";
import KeyValueTable from "../../UI/KeyValueTable/KeyValueTable";
import { CodeEditor } from "../../UI/CodeEditor/CodeEditor";
import { useRecoilValue, RecoilState } from "recoil";
import TrafficViewerApiAtom from "../../../recoil/TrafficViewerApi/atom";
import TrafficViewerApi from "../../TrafficViewer/TrafficViewerApi";
import { toast } from "react-toastify";
import { TOAST_CONTAINER_ID } from "../../../configs/Consts";
import styles from './ReplayRequestModal.module.sass'
import closeIcon from "assets/close.svg"
import spinnerImg from "assets/spinner.svg"


const modalStyle = {
    position: 'absolute',
    top: '6%',
    left: '50%',
    transform: 'translate(-50%, 0%)',
    width: '89vw',
    height: '82vh',
    bgcolor: '#F0F5FF',
    borderRadius: '5px',
    boxShadow: 24,
    p: 4,
    color: '#000',
    padding: "1px 1px",
    paddingBottom: "15px"
};

interface ReplayRequestModalProps {
    isOpen: boolean;
    onClose: () => void;
    request: any
}

enum RequestTabs {
    Params = "params",
    Headers = "headers",
    Body = "body"
}

const isJson = (str) => {
    try {
        JSON.parse(str);
    } catch (e) {
        return false;
    }
    return true;
}

const httpMethods = ['get', 'post', 'put', 'delete']
const TABS = [{ tab: RequestTabs.Params }, { tab: RequestTabs.Headers }, { tab: RequestTabs.Body }];
const convertParamsToArr = (paramsObj) => Object.entries(paramsObj).map(([key, value]) => { return { key, value } })
const ReplayRequestModal: React.FC<ReplayRequestModalProps> = ({ isOpen, onClose, request }) => {

    const [method, setMethod] = useState(request?.method?.toLowerCase() as string)
    const [path, setPath] = useState(request.path);
    const [url, setUrl] = useState("");
    const commonClasses = useCommonStyles();
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);
    const [response, setResponse] = useState(null);
    const [postData, setPostData] = useState(request?.postData?.text || JSON.stringify(request?.postData?.params));
    const [params, setParams] = useState(convertParamsToArr(request?.queryString))
    const [headers, setHeaders] = useState(convertParamsToArr(request?.headers))
    const trafficViewerApi = useRecoilValue(TrafficViewerApiAtom as RecoilState<TrafficViewerApi>)
    const [isLoading, setIsLoading] = useState(false)

    useEffect(() => {
        let newUrl = params.length > 0 ? `${path}?` : path
        params.forEach(({ key, value }) => {
            newUrl += `&${key}=${value}`
        })
        setUrl(newUrl)
    }, [params, path, url])

    const sendRequest = useCallback(async () => {
        const headersData = headers.reduce((prev, corrent) => {
            prev[corrent.key] = corrent.value
            return prev
        }, {})
        const requestData = { url, headers: headersData, data: postData, method }
        try {
            setIsLoading(true)
            const response = await trafficViewerApi.replayRequest(requestData)
            setResponse(response)
        } catch (error) {
            toast.error("Error occurred while fetching response", { containerId: TOAST_CONTAINER_ID });
            console.error(error);
        }
        finally {
            //setIsLoading(false)
        }

    }, [headers, method, postData, trafficViewerApi, url])

    let innerComponent
    switch (currentTab) {
        case RequestTabs.Params:
            innerComponent = <KeyValueTable data={params} onDataChange={(params) => setParams(params)} key={"params"} valuePlaceholder="New Param Value" keyPlaceholder="New param Key" />
            break;
        case RequestTabs.Headers:
            innerComponent = <KeyValueTable data={headers} onDataChange={(heaedrs) => setHeaders(heaedrs)} key={"Header"} valuePlaceholder="New Headers Value" keyPlaceholder="New Headers Key" />
            break;
        case RequestTabs.Body:
            //const formatedCode = formatRequest(postData, request?.postData?.mimeType, true, true, true)
            innerComponent = <div style={{ width: '100%', position: "relative", height: "100%", borderRadius: "inherit" }}>
                <CodeEditor language={request?.postData?.mimeType.split("/")[1]}
                    code={isJson(postData) ? JSON.stringify(JSON.parse(postData || "{}"), null, 2) : postData}
                    onChange={setPostData} />
            </div>
            break;
        default:
            innerComponent = null
            break;
    }

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
                    <div className={styles.headerContainer}>
                        <div className={styles.headerSection}>
                            <span className={styles.title}>Replay Request</span>
                        </div>
                    </div>
                    <div className={styles.modalContainer}>
                        <div className={styles.path}>
                            <select className={styles.select} value={method} onChange={(e) => setMethod(e.target.value)}>
                                {httpMethods.map(method => <option value={method} key={method}>{method}</option>)}
                            </select>
                            <input className={commonClasses.textField} placeholder="Url" value={url}
                                onChange={(event) => setPath(event.target.value)} />
                            <Button size="medium"
                                variant="contained"
                                className={commonClasses.button}
                                onClick={sendRequest}
                                style={{
                                    textTransform: 'uppercase',
                                    width: "fit-content",
                                    marginLeft: "10px"
                                }}>
                                Play
                            </Button >
                        </div>
                        <Tabs tabs={TABS} currentTab={currentTab} onChange={setCurrentTab} leftAligned classes={{ root: styles.tabs }} />
                        <div className={styles.tabContent}>
                            {innerComponent}
                        </div>
                        <div className={styles.responseContainer}>
                            {isLoading && <img alt="spinner" src={spinnerImg} style={{ height: 50 }} />}
                            {response && !isLoading && <SectionsRepresentation data={response} />}
                        </div>
                    </div>
                </Box>
            </Fade>
        </Modal>
    );
}

export default ReplayRequestModal
