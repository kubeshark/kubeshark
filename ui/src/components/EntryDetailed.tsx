import React from "react";
import {singleEntryToHAR} from "../helpers/utils";
import styles from './style/EntryDetailed.module.sass';
import HAREntryViewer from "./HarEntryViewer/HAREntryViewer";
import {makeStyles} from "@material-ui/core";
import StatusCode from "./UI/StatusCode";
import {EndpointPath} from "./UI/EndpointPath";

const useStyles = makeStyles(() => ({
    entryTitle: {
        display: 'flex',
        minHeight: 46,
        maxHeight: 46,
        alignItems: 'center',
        marginBottom: 8,
        padding: 5,
        paddingBottom: 0
    }
}));

interface EntryDetailedProps {
    entryData: any;
    classes?: any;
}

export const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;

const HarEntryTitle: React.FC<any> = ({har}) => {
    const classes = useStyles();

    const {log: {entries}} = har;
    const {response, request, timings: {receive}} = entries[0].entry;
    const {status, statusText, bodySize} = response;


    return <div className={classes.entryTitle}>
        {status && <div style={{marginRight: 8}}>
            <StatusCode statusCode={status}/>
        </div>}
        <div style={{flexGrow: 1, overflow: 'hidden'}}>
            <EndpointPath method={request?.method} path={request?.url}/>
        </div>
        <div style={{margin: "0 18px", opacity: 0.5}}>{formatSize(bodySize)}</div>
        <div style={{marginRight: 18, opacity: 0.5}}>{status} {statusText}</div>
        <div style={{marginRight: 18, opacity: 0.5}}>{Math.round(receive)}ms</div>
        <div style={{opacity: 0.5}}>{'rulesMatched' in entries[0] ? entries[0].rulesMatched?.length : '0'} Rules Applied</div>
    </div>;
};

export const EntryDetailed: React.FC<EntryDetailedProps> = ({classes, entryData}) => {
    const har = singleEntryToHAR(entryData);

    return <>
        {har && <HarEntryTitle har={har}/>}
        <>
            {har && <HAREntryViewer
                harObject={har}
                className={classes?.root ?? styles.har}
            />}
        </>
    </>
};
