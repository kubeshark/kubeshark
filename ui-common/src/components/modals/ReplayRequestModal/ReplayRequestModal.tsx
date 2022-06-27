import { Accordion, AccordionDetails, AccordionSummary, Backdrop, Box, Button, Fade, Modal } from "@mui/material";
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import React, { Fragment, useCallback, useEffect, useMemo, useState } from "react";
import { useCommonStyles } from "../../../helpers/commonStyle";
import { Tabs } from "../../UI";
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
import { formatRequest } from "../../EntryDetailed/EntrySections/EntrySections";
import entryDataAtom from "../../../recoil/entryData";
import { AutoRepresentation } from "../../EntryDetailed/EntryViewer/AutoRepresentation";


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

const httpMethods = ["get", "post", "put", "head", "options", "delete"]
const TABS = [{ tab: RequestTabs.Headers }, { tab: RequestTabs.Params }, { tab: RequestTabs.Body }];
const convertParamsToArr = (paramsObj) => Object.entries(paramsObj).map(([key, value]) => { return { key, value } })
const getQueryStringParams = (link: String) => {
    const query = link ? link.split('?')[1] : ""
    const urlSearchParams = new URLSearchParams(query);
    return Object.fromEntries(urlSearchParams.entries());
};
const ReplayRequestModal: React.FC<ReplayRequestModalProps> = ({ isOpen, onClose }) => {

    const entryData = useRecoilValue(entryDataAtom)
    const request = entryData.data.request
    const [method, setMethod] = useState(request?.method?.toLowerCase() as string)
    const [path, setPath] = useState(request.path);
    const [host, setHost] = useState(entryData.data.dst.name ? entryData.data?.dst?.name : entryData.data.dst.ip)
    const [port, setPort] = useState(entryData.data.dst.port)
    const [finalPath, setFinalPath] = useState("");
    const commonClasses = useCommonStyles();
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);
    const [response, setResponse] = useState(null);
    const [postData, setPostData] = useState(request?.postData?.text || JSON.stringify(request?.postData?.params));
    const [params, setParams] = useState(convertParamsToArr(request?.queryString || {}))
    const [headers, setHeaders] = useState(convertParamsToArr(request?.headers || {}))
    const trafficViewerApi = useRecoilValue(TrafficViewerApiAtom as RecoilState<TrafficViewerApi>)
    const [isLoading, setIsLoading] = useState(false)
    const [requestExpanded, setRequestExpanded] = useState(true)
    const [responseExpanded, setResponseExpanded] = useState(false)


    useEffect(() => {
        //const pathParams = getQueryStringParams(path)
        let newUrl = `${path ? path.split('?')[0] : ""}`
        let finalSpecialChar = ["?", "&", "="].includes(path.slice(-1)) ? path.slice(-1) : ""
        params.forEach(({ key, value }, index) => {
            newUrl += index > 0 ? '&' : '?'
            newUrl += `${key}` + (value ? `=${value}` : "")
        })
        newUrl += finalSpecialChar
        setFinalPath(newUrl)
    // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [params])

    useEffect(() => {
        const newParams = getQueryStringParams(path);
        setParams(convertParamsToArr(newParams))
    }, [path])

    const setHostPort = (hostPort) => {
        const [host, port] = hostPort.split(":")
        setHost(host)
        setPort(port ? port : "")
    }

    const hostPort = useMemo(() => {
        return port ? `${host}:${port}` : host
    }, [host, port])

    const onModalClose = () => {
        setRequestExpanded(true)
        setResponseExpanded(true)
        onClose()
    }

    const sendRequest = useCallback(async () => {
        const headersData = headers.reduce((prev, corrent) => {
            prev[corrent.key] = corrent.value
            return prev
        }, {})
        const buildUrl = `${entryData.base.proto.name}://${hostPort}${path}`
        const requestData = { url: buildUrl, headers: headersData, data: postData, method }
        try {
            setIsLoading(true)
            setRequestExpanded(false)
            setResponseExpanded(true)
            const response = await trafficViewerApi.replayRequest(requestData)
            setResponse(response.data.representation)
            response.errorMessage && toast.error(response.errorMessage, { containerId: TOAST_CONTAINER_ID });
        } catch (error) {
            toast.error("Error occurred while fetching response", { containerId: TOAST_CONTAINER_ID });
            console.error(error);
        }
        finally {
            setIsLoading(false)

        }

    }, [entryData.base.proto.name, headers, hostPort, method, path, postData, trafficViewerApi])

    let innerComponent
    switch (currentTab) {
        case RequestTabs.Params:
            innerComponent = <KeyValueTable data={params} onDataChange={(params) => setParams(params)} key={"params"} valuePlaceholder="New Param Value" keyPlaceholder="New param Key" />
            break;
        case RequestTabs.Headers:
            innerComponent = <Fragment><KeyValueTable data={headers} onDataChange={(heaedrs) => setHeaders(heaedrs)} key={"Header"} valuePlaceholder="New Headers Value" keyPlaceholder="New Headers Key" />
                <span><i>X-Mizu Header added to Request</i></span></Fragment>
            break;
        case RequestTabs.Body:
            const formatedCode = formatRequest(postData, request?.postData?.mimeType)
            innerComponent = <div style={{ width: '100%', position: "relative", height: "100%", borderRadius: "inherit" }}>
                <CodeEditor language={request?.postData?.mimeType.split("/")[1]}
                    code={isJson(formatedCode) ? JSON.stringify(JSON.parse(formatedCode || "{}"), null, 2) : formatedCode}
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
            onClose={onModalClose}
            closeAfterTransition
            BackdropComponent={Backdrop}
            BackdropProps={{ timeout: 500 }}>
            <Fade in={isOpen}>
                <Box sx={modalStyle}>
                    <div className={styles.closeIcon}>
                        <img src={closeIcon} alt="close" onClick={onModalClose} style={{ cursor: "pointer", userSelect: "none" }} />
                    </div>
                    <div className={styles.headerContainer}>
                        <div className={styles.headerSection}>
                            <span className={styles.title}>Replay Request</span>
                        </div>
                    </div>
                    <div className={styles.modalContainer}>
                        <Accordion TransitionProps={{ unmountOnExit: true }} expanded={requestExpanded} onChange={() => setRequestExpanded(!requestExpanded)}>
                            <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="response-content">
                                <span className={styles.sectionHeader}>REQUEST</span>
                            </AccordionSummary>
                            <AccordionDetails>
                                <div className={styles.path}>
                                    <select className={styles.select} value={method} onChange={(e) => setMethod(e.target.value)}>
                                        {httpMethods.map(method => <option value={method} key={method}>{method}</option>)}
                                    </select>
                                    <input placeholder="Host:Port" value={hostPort} onChange={(event) => setHostPort(event.target.value)} className={`${commonClasses.textField} ${styles.hostPort}`} />
                                    <input className={commonClasses.textField} placeholder="Enter Path" value={finalPath}
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
                                        Execute
                                    </Button >
                                </div>
                                <Tabs tabs={TABS} currentTab={currentTab} onChange={setCurrentTab} leftAligned classes={{ root: styles.tabs }} />
                                <div className={styles.tabContent}>
                                    {innerComponent}
                                </div>
                            </AccordionDetails>
                        </Accordion>
                        <Accordion TransitionProps={{ unmountOnExit: true }} expanded={responseExpanded} onChange={() => setResponseExpanded(!responseExpanded)}>
                            <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="response-content">
                                <span className={styles.sectionHeader}>RESPONSE</span>
                            </AccordionSummary>
                            <AccordionDetails>
                                {isLoading && <img alt="spinner" src={spinnerImg} style={{ height: 50 }} />}
                                {response && !isLoading && <AutoRepresentation representation={response} />}
                            </AccordionDetails>
                        </Accordion>
                    </div>
                </Box>
            </Fade>
        </Modal>
    );
}

export default ReplayRequestModal
