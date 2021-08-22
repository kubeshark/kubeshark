import React, {useState} from 'react';
import styles from './HAREntryViewer.module.sass';
import Tabs from "../Tabs";
import {HAREntryTableSection, HAREntryBodySection, HAREntryTablePolicySection} from "./HAREntrySections";

const SectionsRepresentation: React.FC<any> = ({data, color}) => {
    const sections = []

    if (data !== undefined) {
        for (const [i, row] of data.entries()) {
            switch (row.type) {
                case "table":
                    sections.push(
                        <HAREntryTableSection key={i} title={row.title} color={color} arrayToIterate={JSON.parse(row.data)}/>
                    )
                    break;
                case "body":
                    sections.push(
                        <HAREntryBodySection key={i} color={color} content={row.data} encoding={row.encoding} contentType={row.mime_type}/>
                    )
                    break;
                default:
                    break;
            }
        }
    }

    return <>{sections}</>;
}

const AutoRepresentation: React.FC<any> = ({representation, color, isResponseMocked}) => {
    const rulesMatched = []
    const TABS = [
        {tab: 'request'},
        {
            tab: 'response',
            badge: <>{isResponseMocked && <span className="smallBadge virtual mock">MOCK</span>}</>
        },
        {
            tab: 'Rules',
        },
    ];
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);

    // Don't fail even if `representation` is an empty string
    if (representation.length === 0) {
        return <></>;
    }

    const {request, response} = JSON.parse(representation);

    return <div className={styles.harEntry}>
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
                <HAREntryTablePolicySection service={representation.log.entries[0].service} title={'Rule'} color={color} latency={0} response={response} arrayToIterate={rulesMatched ? rulesMatched : []}/>}
            </React.Fragment>}
        </div>}
    </div>;
}

interface Props {
    representation: any;
    color: string,
    isResponseMocked?: boolean;
    showTitle?: boolean;
}

const HAREntryViewer: React.FC<Props> = ({representation, color, isResponseMocked, showTitle=true}) => {
    return <AutoRepresentation representation={representation} color={color} isResponseMocked={isResponseMocked} showTitle={showTitle}/>
};

export default HAREntryViewer;
