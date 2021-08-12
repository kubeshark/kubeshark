import React from "react";
import StatusCode, {getClassification, StatusCodeClassification} from "../UI/StatusCode";
import ingoingIconSuccess from "../assets/ingoing-traffic-success.svg";
import outgoingIconSuccess from "../assets/outgoing-traffic-success.svg";
import ingoingIconFailure from "../assets/ingoing-traffic-failure.svg";
import outgoingIconFailure from "../assets/outgoing-traffic-failure.svg";
import ingoingIconNeutral from "../assets/ingoing-traffic-neutral.svg";
import outgoingIconNeutral from "../assets/outgoing-traffic-neutral.svg";
import styles from "./EntryListItem.module.sass";
import {EndpointPath} from "../UI/EndpointPath";
import {BaseEntry} from "./EntryListItem";

export interface RestEntry extends BaseEntry{
    method?: string,
    path: string,
    service: string,
    statusCode?: number;
    url?: string;
    isCurrentRevision?: boolean;
    isOutgoing?: boolean;
}

interface RestEntryContentProps {
    entry: RestEntry;
}

export const RestEntryContent: React.FC<RestEntryContentProps> = ({entry}) => {
    const classification = getClassification(entry.statusCode)
    let ingoingIcon;
    let outgoingIcon;
    switch (classification) {
        case StatusCodeClassification.SUCCESS: {
            ingoingIcon = ingoingIconSuccess;
            outgoingIcon = outgoingIconSuccess;
            break;
        }
        case StatusCodeClassification.FAILURE: {
            ingoingIcon = ingoingIconFailure;
            outgoingIcon = outgoingIconFailure;
            break;
        }
        case StatusCodeClassification.NEUTRAL: {
            ingoingIcon = ingoingIconNeutral;
            outgoingIcon = outgoingIconNeutral;
            break;
        }
    }
    return <>
        {entry.statusCode && <div>
            <StatusCode statusCode={entry.statusCode}/>
        </div>}
        <div className={styles.endpointServiceContainer}>
            <EndpointPath method={entry.method} path={entry.path}/>
            <div className={styles.service}>
                {entry.service}
            </div>
        </div>
        <div className={styles.directionContainer}>
            {entry.isOutgoing ?
                <img src={outgoingIcon} alt="outgoing traffic" title="outgoing"/>
                :
                <img src={ingoingIcon} alt="ingoing traffic" title="ingoing"/>
            }
        </div>
    </>
}