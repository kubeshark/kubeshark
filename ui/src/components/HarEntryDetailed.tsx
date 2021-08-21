import React from "react";
import {singleEntryToHAR} from "./utils";
import HAREntryViewer from "./HarEntryViewer/HAREntryViewer";
import {makeStyles} from "@material-ui/core";
import Protocol from "./Protocol"
import StatusCode from "./StatusCode";
import {EndpointPath} from "./EndpointPath";

const useStyles = makeStyles(() => ({
    entryTitle: {
        display: 'flex',
        minHeight: 20,
        maxHeight: 46,
        alignItems: 'center',
        marginBottom: 4,
        padding: 2,
        paddingBottom: 0
    },
    entrySummary: {
        display: 'flex',
        minHeight: 36,
        maxHeight: 46,
        alignItems: 'center',
        marginBottom: 4,
        padding: 5,
        paddingBottom: 0
    }
}));

interface HarEntryDetailedProps {
    harEntry: any;
    classes?: any;
}

export const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;

const HarEntryTitle: React.FC<any> = ({protocol, har}) => {
    const classes = useStyles();

    const {log: {entries}} = har;
    const {response} = JSON.parse(entries[0].entry);


    return <div className={classes.entryTitle}>
        <Protocol protocol={protocol} horizontal={true}/>
        <div style={{right: "30px", position: "absolute", display: "flex"}}>
            {response.payload && <div style={{margin: "0 18px", opacity: 0.5}}>{formatSize(response.payload.bodySize)}</div>}
            <div style={{opacity: 0.5}}>{'rulesMatched' in entries[0] ? entries[0].rulesMatched?.length : '0'} Rules Applied</div>
        </div>
    </div>;
};

const HarEntrySummary: React.FC<any> = ({har}) => {
    const classes = useStyles();

    const {log: {entries}} = har;
    const {response, request} = JSON.parse(entries[0].entry);


    return <div className={classes.entrySummary}>
        {response.payload && <div style={{marginRight: 8}}>
            <StatusCode statusCode={response.payload.status}/>
        </div>}
        <div style={{flexGrow: 1, overflow: 'hidden'}}>
            <EndpointPath method={request?.payload.method} path={request?.payload.url}/>
        </div>
    </div>;
};

export const HAREntryDetailed: React.FC<HarEntryDetailedProps> = ({classes, harEntry}) => {
    const har = singleEntryToHAR(harEntry.data);

    return <>
        <HarEntryTitle protocol={harEntry.protocol} har={har}/>
        {har && <HarEntrySummary har={har}/>}
        <>
            {har && <HAREntryViewer representation={harEntry.representation} color={harEntry.protocol.background_color}/>}
        </>
    </>
};
