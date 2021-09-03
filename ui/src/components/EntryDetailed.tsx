import React from "react";
import EntryViewer from "./EntryDetailed/EntryViewer";
import {makeStyles} from "@material-ui/core";
import Protocol from "./UI/Protocol"
import StatusCode from "./UI/StatusCode";
import {EndpointPath} from "./UI/EndpointPath";

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

interface EntryDetailedProps {
    entryData: any
}

export const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;

const EntryTitle: React.FC<any> = ({protocol, data, bodySize, elapsedTime}) => {
    const classes = useStyles();
    const {response} = JSON.parse(data.entry);


    return <div className={classes.entryTitle}>
        <Protocol protocol={protocol} horizontal={true}/>
        <div style={{right: "30px", position: "absolute", display: "flex"}}>
            {response.payload && <div style={{margin: "0 18px", opacity: 0.5}}>{formatSize(bodySize)}</div>}
            <div style={{marginRight: 18, opacity: 0.5}}>{Math.round(elapsedTime)}ms</div>
            <div style={{opacity: 0.5}}>{'rulesMatched' in data ? data.rulesMatched?.length : '0'} Rules Applied</div>
        </div>
    </div>;
};

const EntrySummary: React.FC<any> = ({data}) => {
    const classes = useStyles();

    const {response, request} = JSON.parse(data.entry);

    return <div className={classes.entrySummary}>
        {response?.payload && response.payload?.details && "status" in response.payload.details && <div style={{marginRight: 8}}>
            <StatusCode statusCode={response.payload.details.status}/>
        </div>}
        <div style={{flexGrow: 1, overflow: 'hidden'}}>
            <EndpointPath method={request?.payload.method} path={request?.payload.url}/>
        </div>
    </div>;
};

export const EntryDetailed: React.FC<EntryDetailedProps> = ({entryData}) => {
    return <>
        <EntryTitle
            protocol={entryData.protocol}
            data={entryData.data}
            bodySize={entryData.bodySize}
            elapsedTime={entryData.data.elapsedTime}
        />
        {entryData.data && <EntrySummary data={entryData.data}/>}
        <>
            {entryData.data && <EntryViewer representation={entryData.representation} color={entryData.protocol.background_color}/>}
        </>
    </>
};
