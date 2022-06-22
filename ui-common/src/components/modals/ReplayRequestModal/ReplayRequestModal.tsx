import { Backdrop, Box, Button, Fade, Modal } from "@mui/material";
import React, { useState } from "react";
import styles from './ReplayRequestModal.module.sass'
import closeIcon from "assets/close.svg"
import { useCommonStyles } from "../../../helpers/commonStyle";
import { Tabs } from "../../UI";
import { SectionsRepresentation } from "../../EntryDetailed/EntryViewer/EntryViewer";
import KeyValueTable from "../../UI/KeyValueTable/KeyValueTable";
import CodeEditor from '@uiw/react-textarea-code-editor';


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

const httpMethods = ['get', 'post', 'put', 'delete']
const TABS = [{ tab: RequestTabs.Params }, { tab: RequestTabs.Headers }, { tab: RequestTabs.Body }];
const queryBackgroundColor = "#f5f5f5";
const ReplayRequestModal: React.FC<ReplayRequestModalProps> = ({ isOpen, onClose, request }) => {

    const [selectedMethod, setSelectedMethod] = useState("")
    const [url, setUrl] = useState("");
    const commonClasses = useCommonStyles();
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);
    const [response, setResponse] = useState(null);
    const [postData, setPostData] = useState(request?.postData?.text);

    let innerComponent
    switch (currentTab) {
        case RequestTabs.Params:
            innerComponent = <KeyValueTable data={request.queryString} onDataChange={console.log} />
            break;
        case RequestTabs.Headers:
            innerComponent = <KeyValueTable data={request.headers} onDataChange={console.log} />
            break;
        case RequestTabs.Body:
            innerComponent = <CodeEditor
                value={postData}
                language="js"
                placeholder="request Body"
                onChange={(event) => setPostData(event.target.value)}
                padding={8}
                style={{
                    fontSize: 14,
                    backgroundColor: `${queryBackgroundColor}`,
                    fontFamily: 'ui-monospace,SFMono-Regular,SF Mono,Consolas,Liberation Mono,Menlo,monospace',
                }} />
            break;
        default:
            innerComponent = null
            break;
    }

    const sendRequest = () => { }

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
                            {/* <Button style={{ marginLeft: "2%", textTransform: 'unset' }}
                                startIcon={<img src={refreshIcon} className="custom" alt="refresh"></img>}
                                size="medium"
                                variant="contained"
                                className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
                                onClick={refreshServiceMap}
                            >
                                Refresh
                            </Button> */}
                        </div>
                    </div>
                    <div className={styles.modalContainer}>
                        <div className={styles.path}>
                            <select className={styles.select} value={selectedMethod} onChange={(e) => setSelectedMethod(e.target.value)}>
                                {httpMethods.map(method => <option value={method} key={method}>{method}</option>)}
                            </select>
                            <input className={commonClasses.textField} placeholder="Url" value={url}
                                onChange={(event) => setUrl(event.target.value)} />
                        </div>
                        {/* <div className={styles.requestBody}> */}
                        <Tabs tabs={TABS} currentTab={currentTab} onChange={setCurrentTab} leftAligned />
                        <div className={styles.tabContent}>
                            {innerComponent}
                        </div>
                        {/* </div> */}
                        <Button size="medium"
                            variant="contained"
                            className={commonClasses.button}
                            onClick={sendRequest}
                            style={{ textTransform: 'unset', width: "fit-content" }}>
                            Play
                        </Button >
                        <div className={styles.responseContainer}>
                            {response && <SectionsRepresentation data={response} />}
                        </div>
                    </div>
                </Box>
            </Fade>
        </Modal>
    );
}

export default ReplayRequestModal
