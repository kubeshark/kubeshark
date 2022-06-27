import React, { useState, useCallback } from "react"
import { useRecoilValue } from "recoil"
import entryDataAtom from "../../../recoil/entryData"
import SectionsRepresentation from "./SectionsRepresentation";
import { EntryTablePolicySection, EntryContractSection } from "../EntrySections/EntrySections";
import ReplayRequestModal from "../../modals/ReplayRequestModal/ReplayRequestModal";
import { ReactComponent as ReplayIcon } from './replay.svg';
import styles from './EntryViewer.module.sass';
import { Tabs } from "../../UI";

const enabledProtocolsForReplay = ["http"]

export const AutoRepresentation: React.FC<any> = ({ representation, isRulesEnabled, rulesMatched, contractStatus, requestReason, responseReason, contractContent, elapsedTime, color, isDisplayReplay = false }) => {
    const [isOpenRequestModal, setIsOpenRequestModal] = useState(false)
    const entryData = useRecoilValue(entryDataAtom)
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
    let rulesTabIndex = 0;
    let contractTabIndex = 0;

    if (response) {
        TABS.push(
            {
                tab: 'Response',
                badge: null
            }
        );
        responseTabIndex = TABS.length - 1;
    }

    if (isRulesEnabled) {
        TABS.push(
            {
                tab: 'Rules',
                badge: null
            }
        );
        rulesTabIndex = TABS.length - 1;
    }

    if (contractStatus !== 0 && contractContent) {
        TABS.push(
            {
                tab: 'Contract',
                badge: null
            }
        );
        contractTabIndex = TABS.length - 1;
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
            {isRulesEnabled && currentTab === TABS[rulesTabIndex].tab && <React.Fragment>
                <EntryTablePolicySection title={'Rule'} color={color} latency={elapsedTime} arrayToIterate={rulesMatched ? rulesMatched : []} />
            </React.Fragment>}
            {contractStatus !== 0 && contractContent && currentTab === TABS[contractTabIndex].tab && <React.Fragment>
                <EntryContractSection color={color} requestReason={requestReason} responseReason={responseReason} contractContent={contractContent} />
            </React.Fragment>}
        </div>}
        {isReplayDisplayed() && <ReplayRequestModal request={request} isOpen={isOpenRequestModal} onClose={() => setIsOpenRequestModal(false)} />}
    </div>;
}
