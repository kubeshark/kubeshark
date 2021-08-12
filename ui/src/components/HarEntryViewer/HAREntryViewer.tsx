import React, {useState} from 'react';
import styles from './HAREntryViewer.module.sass';
import Tabs from "../UI/Tabs";
import {TableSection, BodySection, HAREntryTablePolicySection} from "./EntrySections";

const MIME_TYPE_KEY = 'mimeType';

const HAREntryDisplay: React.FC<any> = ({har, entry, isCollapsed: initialIsCollapsed, isResponseMocked}) => {
    const {request, response, timings: {receive}} = entry;
    const rulesMatched = har.log.entries[0].rulesMatched
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

    return <div className={styles.harEntry}>

        {!initialIsCollapsed && <div className={styles.body}>
            <div className={styles.bodyHeader}>
                <Tabs tabs={TABS} currentTab={currentTab} onChange={setCurrentTab} leftAligned/>
                {request?.url && <a className={styles.endpointURL} href={request.url} target='_blank' rel="noreferrer">{request.url}</a>}
            </div>
            {
                currentTab === TABS[0].tab && <React.Fragment>
                    <TableSection title={'Headers'} arrayToIterate={request.headers}/>

                    <TableSection title={'Cookies'} arrayToIterate={request.cookies}/>

                    {request?.postData && <BodySection content={request.postData} encoding={request.postData.comment} contentType={request.postData[MIME_TYPE_KEY]}/>}

                    <TableSection title={'Query'} arrayToIterate={request.queryString}/>
                </React.Fragment>
            }
            {currentTab === TABS[1].tab && <React.Fragment>
                <TableSection title={'Headers'} arrayToIterate={response.headers}/>

                <BodySection content={response.content} encoding={response.content?.encoding} contentType={response.content?.mimeType}/>

                <TableSection title={'Cookies'} arrayToIterate={response.cookies}/>
            </React.Fragment>}
            {currentTab === TABS[2].tab && <React.Fragment>
                <HAREntryTablePolicySection service={har.log.entries[0].service} title={'Rule'} latency={receive} response={response} arrayToIterate={rulesMatched ? rulesMatched : []}/>
            </React.Fragment>}
        </div>}
    </div>;
}

interface Props {
    harObject: any;
    className?: string;
    isResponseMocked?: boolean;
    showTitle?: boolean;
}

const HAREntryViewer: React.FC<Props> = ({harObject, className, isResponseMocked, showTitle=true}) => {
    const {log: {entries}} = harObject;
    const isCollapsed = entries.length > 1;
    return <div className={`${className ? className : ''}`}>
        {Object.keys(entries).map((entry: any, index) => <HAREntryDisplay har={harObject} isCollapsed={isCollapsed} key={index} entry={entries[entry].entry} isResponseMocked={isResponseMocked} showTitle={showTitle}/>)}
    </div>
};

export default HAREntryViewer;
