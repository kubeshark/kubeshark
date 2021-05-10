import React from "react";
import {singleEntryToHAR} from "./utils";
import styles from './style/HarEntryDetailed.module.sass';
import HAREntryViewer from "./HarEntryViewer/HAREntryViewer";
import {makeStyles} from "@material-ui/core";
import StatusCode from "./StatusCode";
import {EndpointPath} from "./EndpointPath";

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

interface HarEntryDetailedProps {
    harEntry: any;
    classes?: any;
}

export const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;

const HarEntryTitle: React.FC<any> = ({har}) => {
    const classes = useStyles();

    const {log: {entries}} = har;
    const {response, request, timings: {receive}} = entries[0];
    const {method, url} = request;
    const {status, statusText, bodySize} = response;


    return <div className={classes.entryTitle}>
        {status && <div style={{marginRight: 8}}>
            <StatusCode statusCode={status}/>
        </div>}
        <div style={{flexGrow: 1, overflow: 'hidden'}}>
            <EndpointPath method={method} path={url}/>
        </div>
        <div style={{margin: "0 24px", opacity: 0.5}}>{formatSize(bodySize)}</div>
        <div style={{marginRight: 24, opacity: 0.5}}>{status} {statusText}</div>
        <div style={{opacity: 0.5}}>{Math.round(receive)}ms</div>
    </div>;
};

export const HAREntryDetailed: React.FC<HarEntryDetailedProps> = ({classes, harEntry}) => {
    const har = singleEntryToHAR(harEntry);

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
