import React, {useState} from 'react';
import styles from './HAREntryViewer.module.sass';
import Tabs from "../Tabs";
import {HAREntryTableSection, HAREntryBodySection} from "./HAREntrySections";

const MIME_TYPE_KEY = 'mimeType';

const HAREntryDisplay: React.FC<any> = ({entry, isCollapsed: initialIsCollapsed, isResponseMocked}) => {
    const {request, response} = entry;

    const TABS = [
        {tab: 'request'},
        {
            tab: 'response',
            badge: <>{isResponseMocked && <span className="smallBadge virtual mock">MOCK</span>}</>
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
                    <HAREntryTableSection title={'Headers'} arrayToIterate={request.headers}/>

                    <HAREntryTableSection title={'Cookies'} arrayToIterate={request.cookies}/>

                    {request?.postData && <HAREntryBodySection content={request.postData} encoding={request.postData.comment} contentType={request.postData[MIME_TYPE_KEY]}/>}

                    <HAREntryTableSection title={'Query'} arrayToIterate={request.queryString}/>
                </React.Fragment>
            }
            {currentTab === TABS[1].tab && <React.Fragment>
                <HAREntryTableSection title={'Headers'} arrayToIterate={response.headers}/>

                <HAREntryBodySection content={response.content} encoding={response.content?.encoding} contentType={response.content?.mimeType}/>

                <HAREntryTableSection title={'Cookies'} arrayToIterate={response.cookies}/>
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
        {Object.keys(entries).map((entry: any, index) => <HAREntryDisplay isCollapsed={isCollapsed} key={index} entry={entries[entry]} isResponseMocked={isResponseMocked} showTitle={showTitle}/>)}
    </div>
};

export default HAREntryViewer;
