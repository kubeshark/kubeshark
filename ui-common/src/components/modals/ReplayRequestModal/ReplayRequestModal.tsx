import { Accordion, AccordionDetails, AccordionSummary, Backdrop, Box, Button, Fade, Modal } from "@mui/material";
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import React, { Fragment, useCallback, useEffect, useState } from "react";
import { useCommonStyles } from "../../../helpers/commonStyle";
import { Tabs } from "../../UI";
import KeyValueTable from "../../UI/KeyValueTable/KeyValueTable";
import CodeEditor from "../../UI/CodeEditor/CodeEditor";
import { useRecoilValue, RecoilState, useRecoilState } from "recoil";
import TrafficViewerApiAtom from "../../../recoil/TrafficViewerApi/atom";
import TrafficViewerApi from "../../TrafficViewer/TrafficViewerApi";
import { toast } from "react-toastify";
import { TOAST_CONTAINER_ID } from "../../../configs/Consts";
import styles from './ReplayRequestModal.module.sass'
import closeIcon from "assets/close.svg"
import refreshImg from "assets/refresh.svg"
import { formatRequestWithOutError } from "../../EntryDetailed/EntrySections/EntrySections";
import entryDataAtom from "../../../recoil/entryData";
import { AutoRepresentation, TabsEnum } from "../../EntryDetailed/EntryViewer/AutoRepresentation";
import useDebounce from "../../../hooks/useDebounce"
import replayRequestModalOpenAtom from "../../../recoil/replayRequestModalOpen";
import { Utils } from "../../../helpers/Utils";
import { LoadingWrapper } from "../../UI/withLoading/withLoading";

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
}

enum RequestTabs {
    Params = "params",
    Headers = "headers",
    Body = "body"
}

const HTTP_METHODS = ["get", "post", "put", "head", "options", "delete"]
const TABS = [{ tab: RequestTabs.Headers }, { tab: RequestTabs.Params }, { tab: RequestTabs.Body }];

const convertParamsToArr = (paramsObj) => Object.entries(paramsObj).map(([key, value]) => { return { key, value } })

const getQueryStringParams = (link: String) => {

    if (link) {
        const decodedURL = decodeQueryParam(link)
        const query = decodedURL.split('?')[1]
        const urlSearchParams = new URLSearchParams(query);
        return Object.fromEntries(urlSearchParams.entries());
    }

    return ""
};

const decodeQueryParam = (p) => {
    return decodeURIComponent(p.replace(/\+/g, ' '));
}

const ReplayRequestModal: React.FC<ReplayRequestModalProps> = ({ isOpen, onClose }) => {
    const entryData = useRecoilValue(entryDataAtom)
    const request = entryData.data.request
    const [method, setMethod] = useState(request?.method?.toLowerCase() as string)
    const getHostUrl = useCallback(() => {
        return entryData.data.dst.name ? entryData.data?.dst?.name : entryData.data.dst.ip
    }, [entryData.data.dst.ip, entryData.data.dst.name])
    const [hostPortInput, setHostPortInput] = useState(`${entryData.base.proto.name}://${getHostUrl()}:${entryData.data.dst.port}`)
    const [pathInput, setPathInput] = useState(request.path);
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

    const debouncedPath = useDebounce(pathInput, 500);

    const onParamsChange = useCallback((newParams) => {
        setParams(newParams);
        let newUrl = `${debouncedPath ? debouncedPath.split('?')[0] : ""}`
        newParams.forEach(({ key, value }, index) => {
            newUrl += index > 0 ? '&' : '?'
            newUrl += `${key}` + (value ? `=${value}` : "")
        })

        setPathInput(newUrl)

    }, [debouncedPath])

    useEffect(() => {
        const newParams = getQueryStringParams(debouncedPath);
        setParams(convertParamsToArr(newParams))
    }, [debouncedPath])

    const onModalClose = () => {
        setRequestExpanded(true)
        setResponseExpanded(true)
        onClose()
    }

    const resetModel = useCallback(() => {
        setMethod(request?.method?.toLowerCase() as string)
        setHostPortInput(`${entryData.base.proto.name}://${getHostUrl()}:${entryData.data.dst.port}`)
        setPathInput(request.path);
        setResponse(null);
        setPostData(request?.postData?.text || JSON.stringify(request?.postData?.params));
        setParams(convertParamsToArr(request?.queryString || {}))
        setHeaders(convertParamsToArr(request?.headers || {}))
        setRequestExpanded(true)
    }, [entryData.base.proto.name, entryData.data.dst.port, getHostUrl, request?.headers, request?.method, request.path, request?.postData?.params, request?.postData?.text, request?.queryString])

    const onRefreshRequest = useCallback((event) => {
        event.stopPropagation()
        resetModel()
    }, [resetModel])


    const sendRequest = useCallback(async () => {
        setResponse(null)
        const headersData = headers.reduce((prev, corrent) => {
            prev[corrent.key] = corrent.value
            return prev
        }, {})
        const buildUrl = `${hostPortInput}${pathInput}`
        const requestData = { url: buildUrl, headers: headersData, data: postData, method }
        try {
            setIsLoading(true)
            const response = await trafficViewerApi.replayRequest(requestData)
            setResponse(response?.data?.representation)
            if (response.errorMessage) {
                toast.error(response.errorMessage, { containerId: TOAST_CONTAINER_ID });
            }
            else {
                setRequestExpanded(false)
                setResponseExpanded(true)
            }

        } catch (error) {
            setRequestExpanded(true)
            toast.error("Error occurred while fetching response", { containerId: TOAST_CONTAINER_ID });
            console.error(error);
        }
        finally {
            setIsLoading(false)
        }

    }, [headers, hostPortInput, method, pathInput, postData, trafficViewerApi])

    let innerComponent
    switch (currentTab) {
        case RequestTabs.Params:
            innerComponent = <div className={styles.keyValueContainer}><KeyValueTable data={params} onDataChange={onParamsChange} key={"params"} valuePlaceholder="New Param Value" keyPlaceholder="New param Key" /></div>
            break;
        case RequestTabs.Headers:
            innerComponent = <Fragment>
                <div className={styles.keyValueContainer}><KeyValueTable data={headers} onDataChange={(heaedrs) => setHeaders(heaedrs)} key={"Header"} valuePlaceholder="New Headers Value" keyPlaceholder="New Headers Key" />
                </div>
                <span className={styles.note}><b>* </b> X-Mizu Header added to reuqests</span>
            </Fragment>
            break;
        case RequestTabs.Body:
            const formatedCode = formatRequestWithOutError(postData || "", request?.postData?.mimeType)
            innerComponent = <div className={styles.codeEditor}>
                <CodeEditor language={request?.postData?.mimeType.split("/")[1]}
                    code={Utils.isJson(formatedCode) ? JSON.stringify(JSON.parse(formatedCode || "{}"), null, 2) : formatedCode}
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
                                <img src={refreshImg} style={{ marginLeft: "10px" }} title="Refresh Reuqest" alt="Refresh Reuqest" onClick={onRefreshRequest} />
                            </AccordionSummary>
                            <AccordionDetails>
                                <div className={styles.path}>
                                    <select className={styles.select} value={method} onChange={(e) => setMethod(e.target.value)}>
                                        {HTTP_METHODS.map(method => <option value={method} key={method}>{method}</option>)}
                                    </select>
                                    <input placeholder="Host:Port" value={hostPortInput} onChange={(event) => setHostPortInput(event.target.value)} className={`${commonClasses.textField} ${styles.hostPort}`} />
                                    <input className={commonClasses.textField} placeholder="Enter Path" value={pathInput}
                                        onChange={(event) => setPathInput(event.target.value)} />
                                    <Button size="medium"
                                        variant="contained"
                                        className={commonClasses.button + ` ${styles.executeButton}`}
                                        onClick={sendRequest}>
                                        Execute
                                    </Button >
                                </div>
                                <Tabs tabs={TABS} currentTab={currentTab} onChange={setCurrentTab} leftAligned classes={{ root: styles.tabs }} />
                                <div className={styles.tabContent}>
                                    {innerComponent}
                                </div>
                            </AccordionDetails>
                        </Accordion>
                        <LoadingWrapper isLoading={isLoading} loaderMargin={10} loaderHeight={50}>
                            {response && (<Accordion TransitionProps={{ unmountOnExit: true }} expanded={responseExpanded} onChange={() => setResponseExpanded(!responseExpanded)}>
                                <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="response-content">
                                    <span className={styles.sectionHeader}>RESPONSE</span>
                                </AccordionSummary>
                                <AccordionDetails>
                                    <AutoRepresentation representation={response} color={entryData.protocol.backgroundColor} openedTab={TabsEnum.Response} />
                                </AccordionDetails>
                            </Accordion>)}
                        </LoadingWrapper>
                    </div>
                </Box>
            </Fade>
        </Modal>
    );
}

const ReplayRequestModalContainer = () => {
    const [isOpenRequestModal, setIsOpenRequestModal] = useRecoilState(replayRequestModalOpenAtom)
    return isOpenRequestModal && < ReplayRequestModal isOpen={isOpenRequestModal} onClose={() => setIsOpenRequestModal(false)} />
}

export default ReplayRequestModalContainer
