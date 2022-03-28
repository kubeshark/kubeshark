import React, {useState} from 'react';
import styles from './EntryViewer.module.sass';
import Tabs from "../../UI/Tabs";
import {EntryTableSection, EntryBodySection, EntryTablePolicySection, EntryContractSection} from "./EntrySections";

enum SectionTypes {
    SectionTable = "table",
    SectionBody = "body",
}

const SectionsRepresentation: React.FC<any> = ({data, color}) => {
    const sections = []

    if (data) {
        for (const [i, row] of data.entries()) {
            switch (row.type) {
                case SectionTypes.SectionTable:
                    sections.push(
                        <EntryTableSection key={i} title={row.title} color={color} arrayToIterate={JSON.parse(row.data)}/>
                    )
                    break;
                case SectionTypes.SectionBody:
                    sections.push(
                        <EntryBodySection key={i} title={row.title} color={color} content={row.data} encoding={row.encoding} contentType={row.mimeType} selector={row.selector}/>
                    )
                    break;
                default:
                    break;
            }
        }
    }

    return <React.Fragment>{sections}</React.Fragment>;
}

const AutoRepresentation: React.FC<any> = ({representation, isRulesEnabled, rulesMatched, contractStatus, requestReason, responseReason, contractContent, elapsedTime, color}) => {
    var TABS = [
        {
            tab: 'Request'
        }
    ];
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);

    // Don't fail even if `representation` is an empty string
    if (!representation) {
        return <React.Fragment></React.Fragment>;
    }

    const {request, response} = JSON.parse(representation);

    let responseTabIndex = 0;
    let rulesTabIndex = 0;
    let contractTabIndex = 0;

    if (response) {
        TABS.push(
            {
                tab: 'Response',
            }
        );
        responseTabIndex = TABS.length - 1;
    }

    if (isRulesEnabled) {
        TABS.push(
            {
                tab: 'Rules',
            }
        );
        rulesTabIndex = TABS.length - 1;
    }

    if (contractStatus !== 0 && contractContent) {
        TABS.push(
            {
                tab: 'Contract',
            }
        );
        contractTabIndex = TABS.length - 1;
    }

    return <div className={styles.Entry}>
        {<div className={styles.body}>
            <div className={styles.bodyHeader}>
                <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned/>
            </div>
            {currentTab === TABS[0].tab && <React.Fragment>
                <SectionsRepresentation data={request} color={color}/>
            </React.Fragment>}
            {response && currentTab === TABS[responseTabIndex].tab && <React.Fragment>
                <SectionsRepresentation data={response} color={color}/>
            </React.Fragment>}
            {isRulesEnabled && currentTab === TABS[rulesTabIndex].tab && <React.Fragment>
                <EntryTablePolicySection title={'Rule'} color={color} latency={elapsedTime} arrayToIterate={rulesMatched ? rulesMatched : []}/>
            </React.Fragment>}
            {contractStatus !== 0 && contractContent && currentTab === TABS[contractTabIndex].tab && <React.Fragment>
                <EntryContractSection color={color} requestReason={requestReason} responseReason={responseReason} contractContent={contractContent}/>
            </React.Fragment>}
        </div>}
    </div>;
}

interface Props {
    representation: any;
    isRulesEnabled: boolean;
    rulesMatched: any;
    contractStatus: number;
    requestReason: string;
    responseReason: string;
    contractContent: string;
    color: string;
    elapsedTime: number;
}

const EntryViewer: React.FC<Props> = ({representation, isRulesEnabled, rulesMatched, contractStatus, requestReason, responseReason, contractContent, elapsedTime, color}) => {
    return <AutoRepresentation
        representation={representation}
        isRulesEnabled={isRulesEnabled}
        rulesMatched={rulesMatched}
        contractStatus={contractStatus}
        requestReason={requestReason}
        responseReason={responseReason}
        contractContent={contractContent}
        elapsedTime={elapsedTime}
        color={color}
    />
};

export default EntryViewer;
