import React, { useState, useCallback } from "react"
import { useRecoilValue, useSetRecoilState } from "recoil"
import entryDataAtom from "../../../recoil/entryData"
import SectionsRepresentation from "./SectionsRepresentation";
import { ReactComponent as ReplayIcon } from './replay.svg';
import styles from './EntryViewer.module.sass';
import { Tabs } from "../../UI";
import replayRequestModalOpenAtom from "../../../recoil/replayRequestModalOpen";

const enabledProtocolsForReplay = ["http"]

export const AutoRepresentation: React.FC<any> = ({ representation, color, isDisplayReplay = false }) => {
    const entryData = useRecoilValue(entryDataAtom)
    const setIsOpenRequestModal = useSetRecoilState(replayRequestModalOpenAtom)
    const isReplayDisplayed = useCallback(() => {
        return enabledProtocolsForReplay.find(x => x === entryData.protocol.name) && isDisplayReplay
    }, [entryData.protocol.name, isDisplayReplay])

    const TABS = [
        {
            tab: 'Request',
            badge: isReplayDisplayed() && <span title="Replay Request"><ReplayIcon fill={color} stroke={color} style={{ marginLeft: "10px", cursor: "pointer", height: "22px" }} onClick={() => setIsOpenRequestModal(true)} /></span>
        }
    ];
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);

    // Don't fail even if `representation` is an empty string
    if (!representation) {
        return <React.Fragment></React.Fragment>;
    }

    const { request, response } = JSON.parse(representation);

    let responseTabIndex = 0;

    if (response) {
        TABS.push(
            {
                tab: 'Response',
                badge: null
            }
        );
        responseTabIndex = TABS.length - 1;
    }

    return <div className={styles.Entry}>
        {<div className={styles.body}>
            <div className={styles.bodyHeader}>
                <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned />
            </div>
            {currentTab === TABS[0].tab && <React.Fragment>
                <SectionsRepresentation data={request} color={color} requestRepresentation={request} />
            </React.Fragment>}
            {response && currentTab === TABS[responseTabIndex].tab && <React.Fragment>
                <SectionsRepresentation data={response} color={color} />
            </React.Fragment>}
        </div>}
    </div>;
}
