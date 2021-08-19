import React from "react";
import {singleEntryToHAR} from "../../../helpers/utils";
import StatusCode from "../../UI/StatusCode";
import {EndpointPath} from "../../UI/EndpointPath";

const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;

export const RestEntryDetailsTitle: React.FC<any> = ({entryData}) => {

    const har = singleEntryToHAR(entryData);
    const {log: {entries}} = har;
    const {response, request, timings: {receive}} = entries[0].entry;
    const {status, statusText, bodySize} = response;

    return har && <>
        {status && <div style={{marginRight: 8}}>
            <StatusCode statusCode={status}/>
        </div>}
        <div style={{flexGrow: 1, overflow: 'hidden'}}>
            <EndpointPath method={request?.method} path={request?.url}/>
        </div>
        <div style={{margin: "0 18px", opacity: 0.5}}>{formatSize(bodySize)}</div>
        <div style={{marginRight: 18, opacity: 0.5}}>{status} {statusText}</div>
        <div style={{marginRight: 18, opacity: 0.5}}>{Math.round(receive)}ms</div>
    </>
}