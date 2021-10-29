import React from "react";
import EntryViewer from "./EntryDetailed/EntryViewer";
import {makeStyles} from "@material-ui/core";
import Protocol from "./UI/Protocol"
import StatusCode from "./UI/StatusCode";
import {Summary} from "./UI/Summary";

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
    updateQuery: any
}

export const formatSize = (n: number) => n > 1000 ? `${Math.round(n / 1000)}KB` : `${n} B`;

const EntryTitle: React.FC<any> = ({protocol, data, bodySize, elapsedTime, updateQuery}) => {
    const classes = useStyles();
    const response = data.response;


    return <div className={classes.entryTitle}>
        <Protocol protocol={protocol} horizontal={true} updateQuery={null}/>
        <div style={{right: "30px", position: "absolute", display: "flex"}}>
            {response && <div
                className="queryable"
                style={{margin: "0 18px", opacity: 0.5}}
                onClick={() => {
                    updateQuery(`response.bodySize == ${bodySize}`)
                }}
            >
                {formatSize(bodySize)}
            </div>}
            {response && <div
                className="queryable"
                style={{marginRight: 18, opacity: 0.5}}
                onClick={() => {
                    updateQuery(`elapsedTime >= ${elapsedTime}`)
                }}
            >
                {Math.round(elapsedTime)}ms
            </div>}
        </div>
    </div>;
};

const EntrySummary: React.FC<any> = ({data, updateQuery}) => {
    const classes = useStyles();

    const response = data.response;

    return <div className={classes.entrySummary}>
        {response && "status" in response && <div style={{marginRight: 8}}>
            <StatusCode statusCode={response.status} updateQuery={updateQuery}/>
        </div>}
        <div style={{flexGrow: 1, overflow: 'hidden'}}>
            <Summary method={data.method} summary={data.summary} updateQuery={updateQuery}/>
        </div>
    </div>;
};

export const EntryDetailed: React.FC<EntryDetailedProps> = ({entryData, updateQuery}) => {
    return <>
        <EntryTitle
            protocol={entryData.protocol}
            data={entryData.data}
            bodySize={entryData.bodySize}
            elapsedTime={entryData.data.elapsedTime}
            updateQuery={updateQuery}
        />
        {entryData.data && <EntrySummary data={entryData.data} updateQuery={updateQuery}/>}
        <>
            {entryData.data && <EntryViewer
                representation={entryData.representation}
                isRulesEnabled={entryData.isRulesEnabled}
                rulesMatched={entryData.rulesMatched}
                contractStatus={entryData.data.contractStatus}
                requestReason={entryData.data.contractRequestReason}
                responseReason={entryData.data.contractResponseReason}
                contractContent={entryData.data.contractContent}
                elapsedTime={entryData.data.elapsedTime}
                color={entryData.protocol.backgroundColor}
                updateQuery={updateQuery}
            />}
        </>
    </>
};
