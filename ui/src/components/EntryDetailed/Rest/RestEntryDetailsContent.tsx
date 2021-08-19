import React, {useState} from "react";
import styles from "../EntryDetailed.module.sass";
import Tabs from "../../UI/Tabs";
import {BodySection, HAREntryTablePolicySection, TableSection} from "../EntrySections";
import {singleEntryToHAR} from "../../../helpers/utils";

const MIME_TYPE_KEY = 'mimeType';

export const RestEntryDetailsContent: React.FC<any> = ({entryData}) => {

    const har = singleEntryToHAR(entryData);
    const {request, response, timings: {receive}} = har.log.entries[0].entry;
    const rulesMatched = har.log.entries[0].rulesMatched
    const TABS = [
        {tab: 'request'},
        {tab: 'response'},
        {tab: 'Rules'},
    ];

    const [currentTab, setCurrentTab] = useState(TABS[0].tab);

    return <>
        <div className={styles.bodyHeader}>
            <Tabs tabs={TABS} currentTab={currentTab} onChange={setCurrentTab} leftAligned/>
            {request?.url && <a className={styles.endpointURL} href={request.url} target='_blank' rel="noreferrer">{request.url}</a>}
        </div>
        {currentTab === TABS[0].tab && <>
                <TableSection title={'Headers'} arrayToIterate={request.headers}/>
                <TableSection title={'Cookies'} arrayToIterate={request.cookies}/>
                {request?.postData && <BodySection content={request.postData} encoding={request.postData.comment} contentType={request.postData[MIME_TYPE_KEY]}/>}
                <TableSection title={'Query'} arrayToIterate={request.queryString}/>
            </>
        }
        {currentTab === TABS[1].tab && <>
            <TableSection title={'Headers'} arrayToIterate={response.headers}/>
            <BodySection content={response.content} encoding={response.content?.encoding} contentType={response.content?.mimeType}/>
            <TableSection title={'Cookies'} arrayToIterate={response.cookies}/>
        </>}
        {currentTab === TABS[2].tab && <>
            <HAREntryTablePolicySection service={har.log.entries[0].service} title={'Rule'} latency={receive} response={response} arrayToIterate={rulesMatched ? rulesMatched : []}/>
        </>}
    </>;
}
