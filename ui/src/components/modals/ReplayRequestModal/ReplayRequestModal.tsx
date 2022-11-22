import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import DownloadIcon from '@mui/icons-material/FileDownloadOutlined';
import UploadIcon from '@mui/icons-material/UploadFile';
import closeIcon from "./assets/close.svg";
import refreshImg from "./assets/refresh.svg";
import { Accordion, AccordionDetails, AccordionSummary, Backdrop, Box, Button, Fade, Modal } from "@mui/material";
import React, { Fragment, useCallback, useEffect, useState } from "react";
import { toast } from "react-toastify";
import { RecoilState, useRecoilState, useRecoilValue } from "recoil";
import { FileContent } from "use-file-picker/dist/interfaces";
import { TOAST_CONTAINER_ID } from "../../../configs/Consts";
import { useCommonStyles } from "../../../helpers/commonStyle";
import { Utils } from "../../../helpers/Utils";
import useDebounce from "../../../hooks/useDebounce";
import entryDataAtom from "../../../recoil/entryData";
import replayRequestModalOpenAtom from "../../../recoil/replayRequestModalOpen";
import TrafficViewerApiAtom from "../../../recoil/TrafficViewerApi/atom";
import { formatRequestWithOutError } from "../../EntryDetailed/EntrySections/EntrySections";
import { AutoRepresentation, TabsEnum } from "../../EntryDetailed/EntryViewer/AutoRepresentation";
import TrafficViewerApi from "../../TrafficViewer/TrafficViewerApi";
import { Tabs } from "../../UI";
import CodeEditor from "../../UI/CodeEditor/CodeEditor";
import FilePicker from '../../UI/FilePicker/FilePicker';
import KeyValueTable, { convertArrToKeyValueObject, convertParamsToArr } from "../../UI/KeyValueTable/KeyValueTable";
import { LoadingWrapper } from "../../UI/withLoading/withLoading";
import { IReplayRequestData, KeyValuePair } from './interfaces';
import styles from './ReplayRequestModal.module.sass';

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

enum RequestTabs {
    Params = "params",
    Headers = "headers",
    Body = "body"
}

const HTTP_METHODS = ["get", "post", "put", "head", "options", "delete"]
const TABS = [{ tab: RequestTabs.Headers }, { tab: RequestTabs.Params }, { tab: RequestTabs.Body }];

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

interface ReplayRequestModalProps {
    isOpen: boolean;
    onClose: () => void;
}

const ReplayRequestModal: React.FC<ReplayRequestModalProps> = ({ isOpen, onClose }) => {
    const entryData = useRecoilValue(entryDataAtom)
    const request = entryData.data.request
    const getHostUrl = useCallback(() => {
        return entryData.data.dst.name ? entryData.data?.dst?.name : entryData.data.dst.ip
    }, [entryData.data.dst.ip, entryData.data.dst.name])
    const getHostPortVal = useCallback(() => {
        return `${entryData.base.proto.name}://${getHostUrl()}:${entryData.data.dst.port}`
    }, [entryData.base.proto.name, entryData.data.dst.port, getHostUrl])
    const [hostPortInput, setHostPortInput] = useState(getHostPortVal())
    const [pathInput, setPathInput] = useState(request.path);
    const commonClasses = useCommonStyles();
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);
    const [response, setResponse] = useState(null);
    const trafficViewerApi = useRecoilValue(TrafficViewerApiAtom as RecoilState<TrafficViewerApi>)
    const [isLoading, setIsLoading] = useState(false)
    const [requestExpanded, setRequestExpanded] = useState(true)
    const [responseExpanded, setResponseExpanded] = useState(false)

    const getInitialRequestData = useCallback((): IReplayRequestData => {
        return {
            method: request?.method?.toLowerCase() as string,
            hostPort: `${entryData.base.proto.name}://${getHostUrl()}:${entryData.data.dst.port}`,
            path: request.path,
            postData: request.postData?.text || JSON.stringify(request.postData?.params),
            headers: convertParamsToArr(request.headers || {}),
            params: convertParamsToArr(request.queryString || {})
        }
    }, [entryData.base.proto.name, entryData.data.dst.port, getHostUrl, request.headers, request?.method, request.path, request.postData?.params, request.postData?.text, request.queryString])

    const [requestDataModel, setRequestData] = useState<IReplayRequestData>(getInitialRequestData())

    const debouncedPath = useDebounce(pathInput, 500);

    const addParamsToUrl = useCallback((url: string, params: KeyValuePair[]) => {
        const urlParams = new URLSearchParams("");
        params.forEach(param => urlParams.append(param.key, param.value as string))
        return `${url}?${urlParams.toString()}`
    }, [])

    const onParamsChange = useCallback((newParams) => {
        let newUrl = `${debouncedPath ? debouncedPath.split('?')[0] : ""}`
        newUrl = addParamsToUrl(newUrl, newParams)
        setPathInput(newUrl)
    }, [addParamsToUrl, debouncedPath])

    useEffect(() => {
        const params = convertParamsToArr(getQueryStringParams(debouncedPath));
        setRequestData({ ...requestDataModel, params })
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [debouncedPath])

    const onModalClose = () => {
        setRequestExpanded(true)
        setResponseExpanded(true)
        onClose()
    }

    const resetModal = useCallback((requestDataModel: IReplayRequestData, hostPortInputVal, pathVal) => {
        setRequestData(requestDataModel)
        setHostPortInput(hostPortInputVal)
        setPathInput(addParamsToUrl(pathVal, requestDataModel.params));
        setResponse(null);
        setRequestExpanded(true);
    }, [addParamsToUrl])

    const onRefreshRequest = useCallback((event) => {
        event.stopPropagation();
        const hostPortInputVal = getHostPortVal();
        resetModal(getInitialRequestData(), hostPortInputVal, request.path);
    }, [getHostPortVal, getInitialRequestData, request.path, resetModal])


    const sendRequest = useCallback(async () => {
        setResponse(null)
        const headersData = convertArrToKeyValueObject(requestDataModel.headers)

        try {
            setIsLoading(true)
            const requestData = { url: `${hostPortInput}${pathInput}`, headers: headersData, data: requestDataModel.postData, method: requestDataModel.method }
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
    }, [hostPortInput, pathInput, requestDataModel.headers, requestDataModel.method, requestDataModel.postData, trafficViewerApi])

    const onDownloadRequest = useCallback((e) => {
        e.stopPropagation()
        const date = Utils.getNow()
        Utils.exportToJson(requestDataModel, `${getHostUrl()} - ${date}`)
    }, [getHostUrl, requestDataModel])

    const onLoadingComplete = useCallback((fileContent: FileContent) => {
        const requestData = JSON.parse(fileContent.content) as IReplayRequestData
        resetModal(requestData, requestData.hostPort, requestData.path)
    }, [resetModal])

    let innerComponent
    switch (currentTab) {
        case RequestTabs.Params:
            innerComponent = <div className={styles.keyValueContainer}><KeyValueTable data={requestDataModel.params} onDataChange={onParamsChange} key={"params"} valuePlaceholder="New Param Value" keyPlaceholder="New param Key" /></div>
            break;
        case RequestTabs.Headers:
            innerComponent = <Fragment>
                <div className={styles.keyValueContainer}><KeyValueTable data={requestDataModel.headers} onDataChange={(headers) => setRequestData({ ...requestDataModel, headers: headers })} key={"Header"} valuePlaceholder="New Headers Value" keyPlaceholder="New Headers Key" />
                </div>
                <span className={styles.note}><b>* </b> X-Kubeshark Header added to requests</span>
            </Fragment>
            break;
        case RequestTabs.Body:
            const formattedCode = formatRequestWithOutError(requestDataModel.postData || "", request?.postData?.mimeType)
            innerComponent = <div className={styles.codeEditor}>
                <CodeEditor language={request?.postData?.mimeType.split("/")[1]}
                    code={Utils.isJson(formattedCode) ? JSON.stringify(JSON.parse(formattedCode || "{}"), null, 2) : formattedCode}
                    onChange={(postData) => setRequestData({ ...requestDataModel, postData })} />
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
                            <Button style={{ marginLeft: "2%", textTransform: 'unset' }}
                                startIcon={<img src={refreshImg} className="custom" alt="Refresh Request"></img>}
                                size="medium"
                                variant="contained"
                                className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
                                onClick={onRefreshRequest}
                            >
                                Refresh
                            </Button>
                            <Button style={{ marginLeft: "2%", textTransform: 'unset' }}
                                startIcon={<DownloadIcon className={`custom ${styles.icon}`} />}
                                size="medium"
                                variant="contained"
                                className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
                                onClick={onDownloadRequest}
                            >
                                Download
                            </Button>
                            <FilePicker onLoadingComplete={onLoadingComplete}
                                elem={<Button style={{ marginLeft: "2%", textTransform: 'unset' }}
                                    startIcon={<UploadIcon className={`custom ${styles.icon}`} />}
                                    size="medium"
                                    variant="contained"
                                    className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}>
                                    Upload
                                </Button>}
                            />
                        </div>
                    </div>
                    <div className={styles.modalContainer}>
                        <Accordion TransitionProps={{ unmountOnExit: true }} expanded={requestExpanded} onChange={() => setRequestExpanded(!requestExpanded)}>
                            <AccordionSummary expandIcon={<ExpandMoreIcon />} aria-controls="response-content">
                                <span className={styles.sectionHeader}>REQUEST</span>
                            </AccordionSummary>
                            <AccordionDetails>
                                <div className={styles.path}>
                                    <select className={styles.select} value={requestDataModel.method} onChange={(e) => setRequestData({ ...requestDataModel, method: e.target.value })}>
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
        </Modal >
    );
}

const ReplayRequestModalContainer = () => {
    const [isOpenRequestModal, setIsOpenRequestModal] = useRecoilState(replayRequestModalOpenAtom)
    return isOpenRequestModal && < ReplayRequestModal isOpen={isOpenRequestModal} onClose={() => setIsOpenRequestModal(false)} />
}

export default ReplayRequestModalContainer
