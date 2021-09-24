import React, {useState} from 'react';
import styles from './EntryViewer.module.sass';
import Tabs from "../UI/Tabs";
import {EntryTableSection, EntryBodySection, EntryTablePolicySection} from "./EntrySections";

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

const AutoRepresentation: React.FC<any> = ({representation, isRulesEnabled, rulesMatched, elapsedTime, color}) => {
    var TABS = [
        {
            tab: 'request'
        },
        {
            tab: 'response',
        },
        {
            tab: 'Rules',
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

    if (!isRulesEnabled) {
        TABS.pop()
    }

    return <div className={styles.Entry}>
        {<div className={styles.body}>
            <div className={styles.bodyHeader}>
                <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned/>
                {request?.url && <a className={styles.endpointURL} href={request.payload.url} target='_blank' rel="noreferrer">{request.payload.url}</a>}
            </div>
            {currentTab === TABS[0].tab && <React.Fragment>
                <SectionsRepresentation data={request} color={color}/>
            </React.Fragment>}
            {response && currentTab === TABS[1].tab && <React.Fragment>
                <SectionsRepresentation data={response} color={color}/>
            </React.Fragment>}
            {TABS.length > 2 && currentTab === TABS[2].tab && <React.Fragment>
                <EntryTablePolicySection title={'Rule'} color={color} latency={elapsedTime} arrayToIterate={rulesMatched ? rulesMatched : []}/>
            </React.Fragment>}
        </div>}
    </div>;
}

interface Props {
    representation: any;
    isRulesEnabled: boolean;
    rulesMatched: any;
    color: string;
    elapsedTime: number;
}

const EntryViewer: React.FC<Props> = ({representation, isRulesEnabled, rulesMatched, elapsedTime, color}) => {
    return <AutoRepresentation representation={representation} isRulesEnabled={isRulesEnabled} rulesMatched={rulesMatched} elapsedTime={elapsedTime} color={color}/>
};

export default EntryViewer;
