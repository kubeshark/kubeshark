import React, {useState} from 'react';
import styles from './EntryViewer.module.sass';
import Tabs from "../UI/Tabs";
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
                        <EntryBodySection key={i} color={color} content={row.data} encoding={row.encoding} contentType={row.mime_type}/>
                    )
                    break;
                default:
                    break;
            }
        }
    }

    return <>{sections}</>;
}

const AutoRepresentation: React.FC<any> = ({representation, isRulesEnabled, rulesMatched, contractStatus, contractReason, elapsedTime, color}) => {
    var TABS = [
        {
            tab: 'Request'
        },
        {
            tab: 'Response',
        },
    ];
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);

    // Don't fail even if `representation` is an empty string
    if (!representation) {
        return <></>;
    }

    const {request, response} = JSON.parse(representation);

    if (!response) {
        TABS[1]['hidden'] = true;
    }

    var rulesTabIndex = 0;
    var contractTabIndex = 0;
    if (isRulesEnabled) {
        TABS.push(
            {
                tab: 'Rules',
            }
        );
        rulesTabIndex = 2;
    }

    if (contractStatus === 2) {
        TABS.push(
            {
                tab: 'Contract',
            }
        );
        if (rulesTabIndex === 0) {
            contractTabIndex = 2;
        } else {
            contractTabIndex = rulesTabIndex + 1;
        }
    }

    return <div className={styles.Entry}>
        {<div className={styles.body}>
            <div className={styles.bodyHeader}>
                <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned/>
            </div>
            {currentTab === TABS[0].tab && <React.Fragment>
                <SectionsRepresentation data={request} color={color}/>
            </React.Fragment>}
            {response && currentTab === TABS[1].tab && <React.Fragment>
                <SectionsRepresentation data={response} color={color}/>
            </React.Fragment>}
            {isRulesEnabled && currentTab === TABS[rulesTabIndex].tab && <React.Fragment>
                <EntryTablePolicySection title={'Rule'} color={color} latency={elapsedTime} arrayToIterate={rulesMatched ? rulesMatched : []}/>
            </React.Fragment>}
            {contractStatus === 2 && currentTab === TABS[contractTabIndex].tab && <React.Fragment>
                <EntryContractSection title={'Contract'} color={color} contractReason={contractReason}/>
            </React.Fragment>}
        </div>}
    </div>;
}

interface Props {
    representation: any;
    isRulesEnabled: boolean;
    rulesMatched: any;
    contractStatus: number;
    contractReason: string;
    color: string;
    elapsedTime: number;
}

const EntryViewer: React.FC<Props> = ({representation, isRulesEnabled, rulesMatched, contractStatus, contractReason, elapsedTime, color}) => {
    return <AutoRepresentation representation={representation} isRulesEnabled={isRulesEnabled} rulesMatched={rulesMatched} contractStatus={contractStatus} contractReason={contractReason} elapsedTime={elapsedTime} color={color}/>
};

export default EntryViewer;
