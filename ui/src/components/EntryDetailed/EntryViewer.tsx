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

const AutoRepresentation: React.FC<any> = ({representation, rulesMatched, elapsedTime, color}) => {
    const TABS = [
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
    return <div className={styles.Entry}>
        {<div className={styles.body}>
            <div className={styles.bodyHeader}>
                <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned/>
                {request?.url && <a className={styles.endpointURL} href={request.payload.url} target='_blank' rel="noreferrer">{request.payload.url}</a>}
            </div>
            {currentTab === TABS[0].tab && <React.Fragment>
                <SectionsRepresentation data={request} color={color}/>
            </React.Fragment>}
            {currentTab === TABS[1].tab && <React.Fragment>
                <SectionsRepresentation data={response} color={color}/>
            </React.Fragment>}
            {currentTab === TABS[2].tab && <React.Fragment>
                {// FIXME: Fix here
                <EntryTablePolicySection service={representation.service} title={'Rule'} color={color} latency={elapsedTime} response={response} arrayToIterate={rulesMatched ? rulesMatched : []}/>}
            </React.Fragment>}
        </div>}
    </div>;
}

interface Props {
    representation: any;
    rulesMatched: any;
    color: string;
    elapsedTime: number;
}

const EntryViewer: React.FC<Props> = ({representation, rulesMatched, elapsedTime, color}) => {
    return <AutoRepresentation representation={representation} rulesMatched={rulesMatched} elapsedTime={elapsedTime} color={color}/>
};

export default EntryViewer;
